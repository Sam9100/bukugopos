package wabot

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// ProcessMessage processes incoming WA message and returns reply
// Supports: simpan, list, hapus, help commands
func ProcessMessage(from string, message string, db *mongo.Database) string {
	msg := strings.TrimSpace(message)
	msgLower := strings.ToLower(msg)

	// Command: help / bantuan
	if msgLower == "help" || msgLower == "bantuan" || msgLower == "menu" {
		return getHelpMessage()
	}

	// Command: simpan / catat
	if strings.HasPrefix(msgLower, "simpan ") || strings.HasPrefix(msgLower, "catat ") {
		keyword := "simpan "
		if strings.HasPrefix(msgLower, "catat ") {
			keyword = "catat "
		}
		content := strings.TrimSpace(msg[len(keyword):])
		return handleSaveNote(from, content, db)
	}

	// Command: list / catatan
	if msgLower == "list" || msgLower == "catatan" || msgLower == "daftar" {
		return handleListNotes(from, db)
	}

	// Command: hapus [nomor]
	if strings.HasPrefix(msgLower, "hapus ") {
		numStr := strings.TrimSpace(msg[len("hapus "):])
		return handleDeleteNote(from, numStr, db)
	}

	// No command matched - return empty to let fallback handler process
	return ""
}

// getHelpMessage returns help menu
func getHelpMessage() string {
	return `🤖 *GOPOS Bot - Menu Bantuan*

📝 *Catatan:*
• simpan [isi] - Simpan catatan baru
• catat [isi] - Sama dengan simpan
• list - Lihat semua catatan
• hapus [nomor] - Hapus catatan

🧠 *AI Assistant (GOPOS AI):*
• Ketik pertanyaan apapun
• AI mengingat 10 pesan terakhir
• Contoh: "ongkir bandung ke jakarta 5kg"

❓ *Bantuan:*
• help - Tampilkan menu ini

💡 Contoh: "simpan beli susu besok"
`
}

// handleSaveNote saves a note for user
func handleSaveNote(userPhone string, content string, db *mongo.Database) string {
	if content == "" {
		return "❌ Isi catatan kosong. Contoh: *simpan beli beras*"
	}

	note := model.Note{
		ID:        primitive.NewObjectID(),
		UserPhone: userPhone,
		Title:     "Catatan",
		Content:   content,
		CreatedAt: time.Now(),
	}

	_, err := atdb.InsertOneDoc(db, "notes", note)
	if err != nil {
		return "❌ Gagal menyimpan: " + err.Error()
	}

	return "✅ Tersimpan!\n\n📝 " + content
}

// handleListNotes returns all notes for user
func handleListNotes(userPhone string, db *mongo.Database) string {
	filter := bson.M{"user_phone": userPhone}
	notes, err := atdb.GetAllDoc[[]model.Note](db, "notes", filter)

	if err != nil {
		return "❌ Gagal mengambil catatan: " + err.Error()
	}

	if len(notes) == 0 {
		return "📭 Belum ada catatan.\n\nKetik *simpan [isi]* untuk menyimpan catatan pertamamu!"
	}

	var sb strings.Builder
	sb.WriteString("📂 *Daftar Catatan:*\n\n")

	for i, note := range notes {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, note.Content))
	}

	sb.WriteString("\n💡 Ketik *hapus [nomor]* untuk menghapus")
	return sb.String()
}

// handleDeleteNote deletes a note by index number
func handleDeleteNote(userPhone string, numStr string, db *mongo.Database) string {
	num, err := strconv.Atoi(strings.TrimSpace(numStr))
	if err != nil || num < 1 {
		return "❌ Format salah. Contoh: *hapus 1*"
	}

	// Get all notes for user
	filter := bson.M{"user_phone": userPhone}
	notes, err := atdb.GetAllDoc[[]model.Note](db, "notes", filter)
	if err != nil {
		return "❌ Gagal: " + err.Error()
	}

	if num > len(notes) {
		return fmt.Sprintf("❌ Catatan nomor %d tidak ditemukan. Kamu punya %d catatan.", num, len(notes))
	}

	// Delete the note at index num-1
	noteToDelete := notes[num-1]
	deleteFilter := bson.M{"_id": noteToDelete.ID}
	_, err = atdb.DeleteOneDoc(db, "notes", deleteFilter)
	if err != nil {
		return "❌ Gagal menghapus: " + err.Error()
	}

	return fmt.Sprintf("🗑️ Catatan #%d dihapus:\n\n~~%s~~", num, noteToDelete.Content)
}
