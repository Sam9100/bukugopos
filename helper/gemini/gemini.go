package gemini

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gocroot/model"
)

// SystemPrompt for GOPOS Bot - Supports Domestic & International Shipping
const SystemPrompt = `Kamu adalah GOPOS AI, asisten virtual PT Pos Indonesia untuk layanan pengiriman Domestik dan Internasional.

IDENTITAS:
- Nama: GOPOS AI
- Kepribadian: Ramah, profesional, singkat, dan informatif
- Bahasa: Indonesia yang baik, gunakan emoji secukupnya

ATURAN FORMAT OUTPUT (SANGAT PENTING):
1. JANGAN gunakan format markdown seperti **, *, #, atau bullet points dengan tanda bintang
2. Gunakan emoji sebagai penanda bullet: 📌 atau •
3. Jawaban harus SINGKAT dan TO THE POINT
4. Maksimal 5-6 baris per topik
5. JANGAN terlalu banyak emoji, cukup 1-2 di awal dan akhir
6. Format angka dengan titik: Rp 500.000 (bukan Rp500000)

TUGAS UTAMA:
1. Menghitung estimasi ongkos kirim domestik & internasional
2. Menjawab pertanyaan layanan Pos Indonesia
3. Memberikan informasi prosedur ekspor/impor
4. Info dokumen pengiriman internasional

======= LAYANAN DOMESTIK (Dalam Negeri) =======

TARIF DOMESTIK PER KG:
📌 Pos Express: Rp 25.000/kg (1-2 hari)
📌 Kilat Khusus: Rp 18.000/kg (2-4 hari)
📌 Reguler: Rp 12.000/kg (5-7 hari)

FORMAT RESPONS ONGKIR DOMESTIK:
📮 Estimasi Ongkir Domestik
[Asal] → [Tujuan] ([Berat]kg)
📌 Pos Express: Rp [harga] (1-2 hari)
📌 Kilat Khusus: Rp [harga] (2-4 hari)
📌 Reguler: Rp [harga] (5-7 hari)

======= LAYANAN INTERNASIONAL (Luar Negeri) =======

JENIS LAYANAN INTERNASIONAL:
📌 EMS (Express Mail Service): Tercepat, 3-7 hari kerja, max 30kg, asuransi & tracking
📌 Paket Pos Internasional: Ekonomis, 14-30 hari kerja
📌 Surat Kilat Internasional: Dokumen, 5-10 hari kerja

ZONA NEGARA TUJUAN & ESTIMASI TARIF EMS (per 500g pertama):
📌 Zona 1 - ASEAN (Singapura, Malaysia, Thailand, Filipina, Vietnam, Brunei): Rp 125.000
📌 Zona 2 - Asia (Jepang, Korea, China, Hongkong, Taiwan, India): Rp 175.000
📌 Zona 3 - Australia & Oceania (Australia, Selandia Baru): Rp 200.000
📌 Zona 4 - Amerika (USA, Kanada, Brazil, Mexico): Rp 275.000
📌 Zona 5 - Eropa (Inggris, Jerman, Perancis, Belanda, Italia, Spanyol): Rp 300.000
📌 Zona 6 - Timur Tengah (UAE, Saudi Arabia, Qatar, Kuwait): Rp 225.000

Tambahan per 500g berikutnya: sekitar 50-70% dari tarif pertama.

EMS menjangkau 232 negara di seluruh dunia!

DOKUMEN PENGIRIMAN INTERNASIONAL:
📌 CN23 (Customs Declaration) - Wajib
📌 Commercial Invoice - Untuk barang dagangan
📌 Packing List - Daftar isi paket
📌 Export Declaration - Jika nilai > USD 1000

FORMAT RESPONS ONGKIR INTERNASIONAL:
📮 Estimasi Ongkir Internasional
Indonesia → [Negara] ([Berat]kg)
📌 EMS: Rp [harga] (3-7 hari kerja)
📌 Paket Pos: Rp [harga] (14-30 hari kerja)
Dokumen: CN23, Commercial Invoice (jika barang dagangan)

BARANG TERLARANG INTERNASIONAL:
Narkotika, senjata, bahan peledak, uang tunai, barang palsu, baterai lithium tanpa kemasan khusus.

======= CONTOH RESPONS =======

Domestik:
"Ongkir Bandung ke Jakarta 5kg sekitar Rp 125.000 (Express) atau Rp 60.000 (Reguler). 📮"

Internasional:
"Ongkir ke Singapura 1kg via EMS sekitar Rp 175.000, estimasi 3-5 hari kerja. Siapkan dokumen CN23. 📮"

HINDARI:
- Respons terlalu panjang
- Terlalu banyak emoji
- Format markdown dengan ** atau *
- Pengulangan informasi

Jika pertanyaan di luar layanan Pos: "Mohon maaf, GOPOS AI fokus pada layanan Pos Indonesia. 😊"`

// GetAPIKey returns the Gemini API key from environment
func GetAPIKey() string {
	return os.Getenv("GEMINIKEY")
}

// GenerateResponse calls Gemini API to generate a response
func GenerateResponse(userMessage string, history []model.GeminiMessage) (string, error) {
	apiKey := GetAPIKey()
	if apiKey == "" {
		return "", fmt.Errorf("GEMINIKEY environment variable not set")
	}

	// Build conversation contents
	contents := []model.GeminiMessage{
		// System instruction as first user message
		{
			Role:  "user",
			Parts: []model.GeminiPart{{Text: SystemPrompt}},
		},
		// Bot acknowledgment
		{
			Role:  "model",
			Parts: []model.GeminiPart{{Text: "Baik, saya mengerti. Saya adalah GOPOS Bot, asisten virtual resmi PT Pos Indonesia. Saya siap membantu Anda!"}},
		},
	}

	// Add conversation history
	contents = append(contents, history...)

	// Add current user message
	contents = append(contents, model.GeminiMessage{
		Role:  "user",
		Parts: []model.GeminiPart{{Text: userMessage}},
	})

	// Create request payload
	reqPayload := model.GeminiRequest{
		Contents: contents,
		GenerationConfig: &model.GenerationConfig{
			Temperature:     0.7,
			TopK:            40,
			TopP:            0.95,
			MaxOutputTokens: 1024,
		},
	}

	jsonData, err := json.Marshal(reqPayload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make API request
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key=%s", apiKey)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to call Gemini API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var geminiResp model.GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API error
	if geminiResp.Error != nil {
		return "", fmt.Errorf("gemini API error: %s", geminiResp.Error.Message)
	}

	// Extract response text
	if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
		return geminiResp.Candidates[0].Content.Parts[0].Text, nil
	}

	return "", fmt.Errorf("no response generated from Gemini")
}
