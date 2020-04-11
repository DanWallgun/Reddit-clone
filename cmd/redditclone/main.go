package main

import (
	"context"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
	"html/template"
	"net/http"
	"os"
	"redditclone/pkg/handlers"
	"redditclone/pkg/middleware"
	"redditclone/pkg/post"
	"redditclone/pkg/session"
	"redditclone/pkg/user"
	"time"

	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	zapLogger, _ := zap.NewProduction()
	defer zapLogger.Sync()
	logger := zapLogger.Sugar()

	// MYSQL DATABASE
	dsn := "root:love@tcp(localhost:3306)/asperitas?"
	dsn += "&charset=utf8"
	dsn += "&interpolateParams=true"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		logger.Errorf("Can't open database. %s", err.Error())
		return
	}
	db.SetMaxOpenConns(5)
	err = db.Ping()
	if err != nil {
		logger.Errorf("Can't ping database. %s", err.Error())
		return
	}
	// DATABASE INIT
	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS asperitas")
	if err != nil {
		logger.Errorf("Can't create database. %s", err.Error())
		return
	}
	_, err = db.Exec(
		"CREATE TABLE IF NOT EXISTS sessions (" +
			"`id` VARCHAR(255) NOT NULL," +
			"`user_id` VARCHAR(255) NOT NULL," +
			"`expires` INT UNSIGNED NOT NULL" +
			");")
	if err != nil {
		logger.Errorf("Can't create table sessions. %s", err.Error())
		return
	}
	_, err = db.Exec(
		"CREATE TABLE IF NOT EXISTS users (" +
			"`id` VARCHAR(255) NOT NULL," +
			"`login` VARCHAR(255) NOT NULL," +
			"`password` VARCHAR(255) NOT NULL" +
			");")
	if err != nil {
		logger.Errorf("Can't create table users. %s", err.Error())
		return
	}

	// MONGO DATABASE
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		logger.Errorf("Can't connect to mongodb. %s", err.Error())
		return
	}
	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		logger.Errorf("MongoDB ping error. %s", err.Error())
		return
	}
	defer func(c *mongo.Client) {
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		err := c.Disconnect(ctx)
		if err != nil {
			logger.Error("Can't disconnect from mongodb. %s", err.Error())
		}
	}(client)
	postsCollection := client.Database("asperitas").Collection("posts")

	// TEMPLATES
	templates := template.Must(template.ParseFiles("./web/index.html"))

	// REPO
	usersRepo := user.NewRepo(db)
	postsRepo := post.NewRepo(postsCollection)
	sessionsManager := session.NewSessionsManager(db)

	// HANDLERS
	usersHandler := handlers.UsersHandler{
		UsersRepo: usersRepo,
		Logger:    logger,
		Sessions:  sessionsManager,
	}

	postsHandler := &handlers.PostsHandler{
		UsersRepo: usersRepo,
		PostsRepo: postsRepo,
		Logger:    logger,
	}

	// ROUTING
	r := mux.NewRouter()
	r.PathPrefix("/static/").Handler(
		http.StripPrefix(
			"/static/",
			http.FileServer(http.Dir("./web/static/")),
		),
	)
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		templates.Execute(w, nil)
	})

	ar := mux.NewRouter()
	authRoute := middleware.Auth(sessionsManager, logger, ar)
	//  API
	//	POST /api/register - регистрация
	r.HandleFunc("/api/register", usersHandler.Register).Methods("POST")
	//	POST /api/login - логин
	r.HandleFunc("/api/login", usersHandler.Login).Methods("POST")
	//	GET /api/posts/ - список всех постов
	r.HandleFunc("/api/posts/", postsHandler.List).Methods("GET")
	//	POST /api/posts/ - добавление поста - обратите внимание - есть с урлом, а есть с текстом
	ar.HandleFunc("/api/posts", postsHandler.Add).Methods("POST")
	r.Handle("/api/posts", authRoute).Methods("POST")
	//	GET /a/funny/{CATEGORY_NAME} - список постов конкретной категории
	r.HandleFunc("/api/posts/{category_name}", postsHandler.ListByCategory).Methods("GET")
	//	GET /api/post/{POST_ID} - детали поста с комментами
	r.HandleFunc("/api/post/{post_id}", postsHandler.GetPost).Methods("GET")
	//	POST /api/post/{POST_ID} - добавление коммента
	ar.HandleFunc("/api/post/{post_id}", postsHandler.Comment).Methods("POST")
	r.Handle("/api/post/{post_id}", authRoute).Methods("POST")
	//	DELETE /api/post/{POST_ID}/{COMMENT_ID} - удаление коммента
	ar.HandleFunc("/api/post/{post_id}/{comment_id}", postsHandler.DeleteComment).Methods("DELETE")
	r.Handle("/api/post/{post_id}/{comment_id}", authRoute).Methods("DELETE")
	//	GET /api/post/{POST_ID}/upvote - рейтинг постп вверх
	ar.HandleFunc("/api/post/{post_id}/upvote", postsHandler.Upvote).Methods("GET")
	r.Handle("/api/post/{post_id}/upvote", authRoute).Methods("GET")
	//	GET /api/post/{POST_ID}/downvote - рейтинг поста вниз
	ar.HandleFunc("/api/post/{post_id}/downvote", postsHandler.Downvote).Methods("GET")
	r.Handle("/api/post/{post_id}/downvote", authRoute).Methods("GET")
	//	DELETE /api/post/{POST_ID} - удаление поста
	ar.HandleFunc("/api/post/{post_id}", postsHandler.DeletePost).Methods("DELETE")
	r.Handle("/api/post/{post_id}", authRoute).Methods("DELETE")
	//	GET /api/user/{USER_LOGIN} - получение всех постов конкртеного пользователя
	r.HandleFunc("/api/user/{user_login}", postsHandler.ListByAuthor).Methods("GET")
	// unvoting
	ar.HandleFunc("/api/post/{post_id}/unvote", postsHandler.Unvote).Methods("GET")
	r.Handle("/api/post/{post_id}/unvote", authRoute).Methods("GET")

	handler := middleware.AccessLog(logger, r)
	handler = middleware.Panic(handler)

	// LET'S START A WAR
	port := os.Getenv("PORT")
	err = http.ListenAndServe(":"+port, handler)
	if err != nil {
		logger.Error(err.Error())
	}
}
