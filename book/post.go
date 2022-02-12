package book

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"fud_library/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreatePost(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")

	var bookPost Post

	err := utils.ParseJSONFromRequest(request, &bookPost)

	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, response)
		return
	}

	blogTitle := strings.ToTitle(bookPost.Title)

	// confirm if blog title has already been taken
	result, _ := utils.GetMongoDBDoc(PostCollectionName, bson.M{"title": blogTitle})

	if result != nil {
		utils.GetError(
			fmt.Errorf(fmt.Sprintf("blog post with title %s exists!", blogTitle)),
			http.StatusBadRequest,
			response,
		)

		return
	}

	bookPost.Title = blogTitle
	bookPost.Deleted = false
	bookPost.Likes = 0
	bookPost.Comments = 0
	bookPost.Length = calculateReadingTime(bookPost.Content)
	bookPost.CreatedAt = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), time.Now().UTC().Hour(), time.Now().Minute(), time.Now().Second(), 0, time.Local)

	detail, _ := utils.StructToMap(bookPost)

	res, err := utils.CreateMongoDBDoc(PostCollectionName, detail)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, response)
		return
	}

	insertedPostID := res.InsertedID.(primitive.ObjectID).Hex()

	blogPostLikes := Likes{ID: insertedPostID, UsersList: []string{}}
	blogPostLikesMap, _ := utils.StructToMap(blogPostLikes)
	likeDocResponse, err := utils.CreateMongoDBDoc(PostLikesCollectionName, blogPostLikesMap)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, response)
		return
	}

	blogPostComments := BlogsComment{ID: insertedPostID, Comments: []Comment{}}
	blogPostCommentsMap, _ := utils.StructToMap(blogPostComments)

	commentDocResponse, err := utils.CreateMongoDBDoc(PostCommentsCollectionName, blogPostCommentsMap)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, response)
		return
	}

	ress := []interface{}{res, likeDocResponse, commentDocResponse}

	_ = ress

	utils.GetSuccess("Post created", bson.M{}, response)
}

// An endpoint to list all available posts.
func GetPosts(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")

	blogs, err := utils.GetMongoDBDocs(PostCollectionName, bson.M{"deleted": false})
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, response)
		return
	}

	utils.GetSuccess("success", blogs, response)
}

func calculateReadingTime(content string) int {
	words := strings.Split(content, " ")
	wordLength := len(words)
	readingTime := wordLength / 200

	return readingTime
}
