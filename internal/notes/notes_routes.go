package notes

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func RegisterRoutes(r *gin.Engine, db *mongo.Database) {
	repo := NewRepo(db)
	handler := NewHandler(repo)

	notesGroup := r.Group("/notes")
	{
		notesGroup.POST("/add", handler.createNote)
		notesGroup.GET("", handler.ListNotes)
		notesGroup.GET("id/:id", handler.GetNoteByID)
		notesGroup.PUT("id/:id", handler.UpdateNotebyID)

		notesGroup.DELETE("id/:id", handler.DeleteNoteByID)
	}
}
