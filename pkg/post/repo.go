package post

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"redditclone/pkg/user"
	"redditclone/pkg/utils"
	"redditclone/pkg/wrap"
	"time"
)

type PostsRepo struct {
	Posts wrap.CollectionHelper
}

func NewRepo(collection wrap.CollectionHelper) *PostsRepo {
	return &PostsRepo{
		Posts: collection,
	}
}

var (
	ErrNoPost = errors.New("No post found")
)

func (repo *PostsRepo) GetByFilter(filter interface{}) ([]*Post, error) {
	posts := []*Post{}
	//ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	ctx := context.Background()
	cur, err := repo.Posts.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var result Post
		err := cur.Decode(&result)
		if err != nil {
			return nil, err
		}
		posts = append(posts, &result)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}

func (repo *PostsRepo) GetAll() ([]*Post, error) {
	filter := bson.M{}
	return repo.GetByFilter(filter)
}

func (repo *PostsRepo) GetByCategory(category string) ([]*Post, error) {
	filter := bson.M{"category": category}
	return repo.GetByFilter(filter)
}

func (repo *PostsRepo) GetByID(id string) (*Post, error) {
	filter := bson.M{"id": id}
	posts, err := repo.GetByFilter(filter)
	if err != nil {
		return nil, err
	}
	return posts[0], nil
}

func (repo *PostsRepo) GetByAuthor(login string) ([]*Post, error) {
	filter := bson.M{"author.login": login}
	return repo.GetByFilter(filter)
}

func (repo *PostsRepo) Add(post *Post) (*Post, error) {
	id := utils.GenerateRandomString(32)
	post.ID = id
	post.Votes = []Vote{
		{
			User: post.Author.ID,
			Vote: 1,
		},
	}
	post.Comments = make([]Comment, 0)
	post.UpvotePercentage = 100
	post.Views = 0
	post.Score = 1
	post.Created = time.Now()

	//ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	ctx := context.Background()
	_, err := repo.Posts.InsertOne(ctx, post)
	if err != nil {
		return nil, err
	}

	return post, nil
}

func (repo *PostsRepo) Comment(postID string, u *user.User, comment string) (*Post, error) {
	id := utils.GenerateRandomString(32)

	filter := bson.M{"id": postID}
	post, err := repo.GetByID(postID)
	if err != nil {
		return nil, ErrNoPost
	}
	post.Comments = append(post.Comments, Comment{
		Created: time.Now(),
		Author:  u,
		Body:    comment,
		ID:      id,
	})

	ctx := context.Background()
	//ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	_, err = repo.Posts.ReplaceOne(ctx, filter, post)
	if err != nil {
		return nil, err
	}

	return post, nil
}

func (repo *PostsRepo) DeleteComment(postID, commentID string) (*Post, error) {
	filter := bson.M{"id": postID}
	post, err := repo.GetByID(postID)
	if err != nil {
		return nil, ErrNoPost
	}
	for i, c := range post.Comments {
		if c.ID == commentID {
			post.Comments[i] = post.Comments[len(post.Comments)-1]
			post.Comments = post.Comments[:len(post.Comments)-1]
			break
		}
	}
	//ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	ctx := context.Background()
	_, err = repo.Posts.ReplaceOne(ctx, filter, post)
	if err != nil {
		return nil, err
	}

	return post, nil
}

func (repo *PostsRepo) DeletePost(postID string) error {
	filter := bson.M{"id": postID}
	//ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	ctx := context.Background()
	_, err := repo.Posts.DeleteOne(ctx, filter)
	if err != nil {
		return ErrNoPost
	}
	return nil
}

func (repo *PostsRepo) Vote(postID, userID string, v int32) (*Post, error) {
	filter := bson.M{"id": postID}
	post, err := repo.GetByID(postID)
	if err != nil {
		return nil, ErrNoPost
	}
	flag := true
	for i, vote := range post.Votes {
		if vote.User == userID {
			post.Votes[i].Vote = v
			recount(post)
			flag = false
			break
		}
	}
	if flag {
		post.Votes = append(post.Votes, Vote{
			User: userID,
			Vote: v,
		})
		recount(post)
	}

	//ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	ctx := context.Background()
	_, err = repo.Posts.ReplaceOne(ctx, filter, post)
	if err != nil {
		return nil, err
	}

	return post, nil
}

func (repo *PostsRepo) Upvote(postID, userID string) (*Post, error) {
	return repo.Vote(postID, userID, 1)
}

func (repo *PostsRepo) Downvote(postID, userID string) (*Post, error) {
	return repo.Vote(postID, userID, -1)
}

func (repo *PostsRepo) Unvote(postID, userID string) (*Post, error) {
	return repo.Vote(postID, userID, 0)
}

func recount(post *Post) {
	post.Score = 0
	up := 0
	for _, v := range post.Votes {
		post.Score += v.Vote
		if v.Vote == 1 {
			up += 1
		}
	}
	post.UpvotePercentage = uint32(up * 100 / len(post.Votes))
}
