package wagemini

import (
	"context"
	"fmt"
	"time"

	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/gemini"
	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// MaxHistoryMessages - Maximum messages to keep in history for context
	MaxHistoryMessages = 10
	// CollectionName - MongoDB collection for WA chat history
	CollectionName = "wa_chat_history"
)

// GetAIResponse generates AI response with conversation history context
func GetAIResponse(phoneNumber, message string, db *mongo.Database) (string, error) {
	// Get conversation history
	history, err := getHistory(phoneNumber, db)
	if err != nil {
		fmt.Printf("⚠️ Failed to get history for %s: %v\n", phoneNumber, err)
		history = []model.GeminiMessage{} // Continue without history
	}

	// Generate AI response using existing gemini helper
	aiResponse, err := gemini.GenerateResponse(message, history)
	if err != nil {
		return "", fmt.Errorf("gemini error: %w", err)
	}

	// Save conversation to history
	if saveErr := saveMessage(phoneNumber, message, aiResponse, db); saveErr != nil {
		fmt.Printf("⚠️ Failed to save history: %v\n", saveErr)
		// Don't fail - still return the response
	}

	return aiResponse, nil
}

// getHistory retrieves conversation history from MongoDB and converts to GeminiMessage format
func getHistory(phoneNumber string, db *mongo.Database) ([]model.GeminiMessage, error) {
	filter := bson.M{"phone_number": phoneNumber}
	chatHistory, err := atdb.GetOneDoc[model.WAChatHistory](db, CollectionName, filter)
	if err != nil {
		// No history found - return empty slice
		return []model.GeminiMessage{}, nil
	}

	// Convert WAChatMessage to GeminiMessage format
	var geminiHistory []model.GeminiMessage
	for _, msg := range chatHistory.Messages {
		geminiHistory = append(geminiHistory, model.GeminiMessage{
			Role: msg.Role,
			Parts: []model.GeminiPart{
				{Text: msg.Content},
			},
		})
	}

	return geminiHistory, nil
}

// saveMessage saves user message and AI response to MongoDB
func saveMessage(phoneNumber, userMessage, aiResponse string, db *mongo.Database) error {
	now := time.Now()

	// Create new messages
	newMessages := []model.WAChatMessage{
		{
			Role:      "user",
			Content:   userMessage,
			Timestamp: now,
		},
		{
			Role:      "model",
			Content:   aiResponse,
			Timestamp: now,
		},
	}

	// Try to find existing history
	filter := bson.M{"phone_number": phoneNumber}
	existingHistory, err := atdb.GetOneDoc[model.WAChatHistory](db, CollectionName, filter)

	if err != nil {
		// No existing history - create new document
		newHistory := model.WAChatHistory{
			ID:          primitive.NewObjectID(),
			PhoneNumber: phoneNumber,
			Messages:    newMessages,
			UpdatedAt:   now,
		}
		_, insertErr := atdb.InsertOneDoc(db, CollectionName, newHistory)
		return insertErr
	}

	// Append new messages to existing history
	allMessages := append(existingHistory.Messages, newMessages...)

	// Trim to max messages (keep most recent)
	if len(allMessages) > MaxHistoryMessages {
		allMessages = allMessages[len(allMessages)-MaxHistoryMessages:]
	}

	// Update document
	update := bson.M{
		"$set": bson.M{
			"messages":   allMessages,
			"updated_at": now,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err = db.Collection(CollectionName).UpdateOne(context.Background(), filter, update, opts)
	return err
}

// ClearHistory clears conversation history for a phone number
func ClearHistory(phoneNumber string, db *mongo.Database) error {
	filter := bson.M{"phone_number": phoneNumber}
	_, err := db.Collection(CollectionName).DeleteOne(context.Background(), filter)
	return err
}
