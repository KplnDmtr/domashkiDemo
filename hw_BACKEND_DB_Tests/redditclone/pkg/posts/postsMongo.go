package posts

import (
	"context"
	"fmt"
	"slices"

	"redditclone/pkg/comments"
	"redditclone/pkg/vote"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PostsMongoRepo struct {
	Posts *mongo.Collection
}

func NewPostsMongoRepo(collection *mongo.Collection) *PostsMongoRepo {
	repo := &PostsMongoRepo{
		Posts: collection,
	}
	return repo
}

func (p *PostsMongoRepo) GetAllPosts(ctx context.Context) ([]Post, error) {
	posts := make([]Post, 0)

	c, err := p.Posts.Find(ctx, bson.M{})
	if err != nil {
		return posts, fmt.Errorf("error in getallposts:%s", err.Error())
	}
	defer c.Close(ctx)
	err = c.All(ctx, &posts)
	if err != nil {
		return posts, fmt.Errorf("error in getallposts:%s", err.Error())
	}
	return posts, nil
}

func (p *PostsMongoRepo) AddPost(ctx context.Context, post Post) (Post, error) {
	post.ID = primitive.NewObjectID().Hex()
	post.Upvotes++
	newPost, err := bson.Marshal(post)
	if err != nil {
		return Post{}, fmt.Errorf("error in adpost: %s", err.Error())
	}

	res, err := p.Posts.InsertOne(ctx, newPost)
	if err != nil {
		return Post{}, fmt.Errorf("error in adpost: %s", err.Error())
	}
	post.ID = res.InsertedID.(string)
	return post, nil
}

func (p *PostsMongoRepo) UpdatePost(ctx context.Context, post Post) error {
	updatePost := bson.M{
		"$set": post,
	}
	res, err := p.Posts.UpdateByID(ctx, post.ID, updatePost)
	if err != nil {
		return fmt.Errorf("error in updpost: %s", err.Error())
	}
	if res.ModifiedCount == 0 {
		return fmt.Errorf("updatepost - not found what to modify")
	}
	return nil
}

func (p *PostsMongoRepo) GetPostByID(ctx context.Context, id string) (Post, error) {
	post := Post{}
	res := p.Posts.FindOne(ctx, bson.M{"_id": id})
	if res.Err() != nil {
		return Post{}, fmt.Errorf("error in getpost: %s", res.Err().Error())
	}
	err := res.Decode(&post)
	if err != nil {
		return Post{}, fmt.Errorf("error in getpost: %s", err.Error())
	}
	return post, nil
}

func (p *PostsMongoRepo) GetCategory(ctx context.Context, category string) ([]Post, error) {
	posts := make([]Post, 0)

	c, err := p.Posts.Find(ctx, bson.M{"category": category})
	if err != nil {
		return posts, fmt.Errorf("error in getallposts:%s", err.Error())
	}
	defer c.Close(ctx)

	err = c.All(ctx, &posts)
	if err != nil {
		return posts, fmt.Errorf("error in getallposts:%s", err.Error())
	}
	return posts, nil
}

func (p *PostsMongoRepo) DeletePost(ctx context.Context, postID string) error {
	res, err := p.Posts.DeleteOne(ctx, bson.M{"_id": postID})
	if err != nil {
		return fmt.Errorf("error in deletePost %s", err.Error())
	}
	if res.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (p *PostsMongoRepo) AddComment(ctx context.Context, postID string, comment comments.Comment) (Post, error) {
	comment.ID = primitive.NewObjectID().Hex()
	update := bson.M{
		"$push": bson.M{
			"comments": comment,
		},
	}
	options := options.FindOneAndUpdate().SetReturnDocument(options.After)
	res := p.Posts.FindOneAndUpdate(ctx, bson.M{"_id": postID}, update, options)
	if res.Err() != nil {
		return Post{}, fmt.Errorf("error in addcomment: %s", res.Err().Error())
	}
	post := Post{}
	err := res.Decode(&post)
	if err != nil {
		return Post{}, fmt.Errorf("error in addcomment: %s", err.Error())
	}

	return post, nil
}

func (p *PostsMongoRepo) DeleteComment(ctx context.Context, postID string, commentID string) (Post, error) {

	filter := bson.M{
		"_id": postID,
	}
	update := bson.M{
		"$pull": bson.M{"comments": bson.M{"id": commentID}},
	}
	options := options.FindOneAndUpdate().SetReturnDocument(options.After)
	res := p.Posts.FindOneAndUpdate(ctx, filter, update, options)
	if res.Err() != nil {
		return Post{}, fmt.Errorf("error in DeleteComment: %s", res.Err().Error())
	}
	post := Post{}
	err := res.Decode(&post)
	if err != nil {
		return Post{}, fmt.Errorf("error in DeleteComment: %s", err.Error())
	}

	return post, nil

}

func (p *PostsMongoRepo) GetByUserLogin(ctx context.Context, login string) ([]Post, error) {
	posts := make([]Post, 0)

	c, err := p.Posts.Find(ctx, bson.M{"author.username": login})
	if err != nil {
		return posts, fmt.Errorf("error in geByUserLoginPosts:%s", err.Error())
	}
	defer c.Close(ctx)

	err = c.All(ctx, &posts)
	if err != nil {
		return posts, fmt.Errorf("error in getallposts:%s", err.Error())
	}
	return posts, nil
}

func (p *PostsMongoRepo) findVote(votes []vote.Vote, creator string) (int, bool) {
	for index, voteTmp := range votes {
		if voteTmp.User == creator {
			return index, true
		}
	}
	return -1, false
}

func (p *PostsMongoRepo) addVote(post *Post, vote vote.Vote) {
	post.Votes = append(post.Votes, vote)
	if vote.Vote > 0 {
		post.Upvotes++
	}
	post.Score += vote.Vote
}

func (p *PostsMongoRepo) updateVote(post *Post, index int, vote vote.Vote) {
	post.Score -= post.Votes[index].Vote
	post.Votes[index].Vote = vote.Vote
	post.Upvotes += vote.Vote
	post.Score += vote.Vote
}

func (p *PostsMongoRepo) updateVoteStats(post *Post) {
	if len(post.Votes) != 0 {
		post.UpvotePercentage = int((float32(post.Upvotes) / float32(len(post.Votes)) * 100))
	} else {
		post.UpvotePercentage = 0
	}
}

func (p *PostsMongoRepo) deleteVote(post *Post, index int) {
	if len(post.Votes) != 0 {
		post.Votes = slices.Delete(post.Votes, index, index+1)
	}
}

func (p *PostsMongoRepo) Vote(ctx context.Context, postID string, vote vote.Vote) (Post, error) {
	post, err := p.GetPostByID(ctx, postID)
	if err != nil {
		return Post{}, err
	}

	if index, voteExist := p.findVote(post.Votes, vote.User); !voteExist {
		p.addVote(&post, vote)
	} else {
		p.updateVote(&post, index, vote)
	}

	p.updateVoteStats(&post)

	err = p.UpdatePost(ctx, post)
	if err != nil {
		return Post{}, err
	}

	return post, nil
}

func (p *PostsMongoRepo) UnVote(ctx context.Context, username string, postID string) (Post, error) {
	post, err := p.GetPostByID(ctx, postID)
	if err != nil {
		return Post{}, err
	}

	var index int
	var voteExist bool
	if index, voteExist = p.findVote(post.Votes, username); !voteExist {
		return Post{}, fmt.Errorf("no such vote")
	}

	post.Score -= post.Votes[index].Vote
	if post.Votes[index].Vote > 0 && post.Upvotes > 0 {
		post.Upvotes--
	}
	p.deleteVote(&post, index)
	p.updateVoteStats(&post)

	err = p.UpdatePost(ctx, post)
	if err != nil {
		return Post{}, err
	}

	return post, nil
}
