package notes

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Handler struct {
	repo *Repo
}

func NewHandler(repo *Repo) *Handler {
	return &Handler{
		repo: repo,
	}
}

func (h *Handler) createNote(ctx *gin.Context) {
	var req CreateNoteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	now := time.Now().UTC()
	note := &Note{
		ID:        primitive.NewObjectID(),
		Title:     req.Title,
		Content:   req.Content,
		CreatedAt: now,
		UpdatedAt: now,
		Pinned:    req.Pinned,
	}

	created, err := h.repo.CreateNote(ctx.Request.Context(), note)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create note"})
		return
	}

	ctx.JSON(http.StatusCreated, created)

}

func (h *Handler) ListNotes(c *gin.Context) {
	notes, err := h.repo.List(c.Request.Context()) // or GetAllNotes, whichever repo method you prefer
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch notes",
		})
		return
	}

	// Ensure we send an empty array [] instead of null if no notes exist
	if notes == nil {
		notes = []Note{}
	}

	// Wrapped in an object for future scalability (pagination, etc.)
	c.JSON(http.StatusOK, gin.H{
		"notes": notes,
	})
}
func (h *Handler) DeleteNoteByID(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	err = h.repo.DeleteNotebyID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "No note found with that ID",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete note",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Note successfully deleted",
		"id":      id.Hex(),
	})
}
func (h *Handler) GetNoteByID(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	note, err := h.repo.GetNoteByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "no ID matches that description",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch note",
		})
		return
	}

	c.JSON(http.StatusOK, note)
}

func (h *Handler) UpdateNotebyID(c *gin.Context) {
	idstr := c.Param("id")

	objID, err := primitive.ObjectIDFromHex(idstr)

	if err != nil {
		c.JSON(
			http.StatusBadRequest, gin.H{
				"error": "invalid ID",
			},
		)
		return
	}

	var req UpdateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(
			http.StatusBadRequest, gin.H{
				"error": "invalid json format",
			},
		)
		return
	}

	updated, err := h.repo.UpdateNote(c.Request.Context(), objID, req)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "no ID matches that description",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch note",
		})
		return
	}
	c.JSON(http.StatusOK, updated)
}
