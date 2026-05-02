package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type AnalyzeReceiptResponse struct {
	Date   string `json:"date"`
	Shop   string `json:"shop"`
	Item   string `json:"item"`
	Amount int    `json:"amount"`
}

func AnalyzeReceipt(c *gin.Context) {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: GOOGLE_API_KEY is not set")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "GOOGLE_API_KEY is not set"})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		fmt.Printf("Error: Image is required: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image is required"})
		return
	}

	src, err := file.Open()
	if err != nil {
		fmt.Printf("Error: Failed to open image: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open image"})
		return
	}
	defer src.Close()

	imgData, err := io.ReadAll(src)
	if err != nil {
		fmt.Printf("Error: Failed to read image: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read image"})
		return
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		fmt.Printf("Error: Failed to create Gemini client: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Gemini client"})
		return
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-flash-latest")

	prompt := []genai.Part{
		genai.ImageData("jpeg", imgData),
		genai.Text("Analyze this receipt and return JSON only. Use YYYY-MM-DD for date, name for shop, summary for item, and integer for amount. JSON:\n{\"date\": \"YYYY-MM-DD\", \"shop\": \"name\", \"item\": \"summary\", \"amount\": 1234}"),
	}

	resp, err := model.GenerateContent(ctx, prompt...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate content: %v", err)})
		return
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No results from Gemini"})
		return
	}

	// レスポンスからJSONを抽出
	var resultText string
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			resultText += string(text)
		}
	}

	// Markdownのコードブロックが含まれている場合があるので除去
	resultText = strings.TrimSpace(resultText)
	resultText = strings.TrimPrefix(resultText, "```json")
	resultText = strings.TrimPrefix(resultText, "```")
	resultText = strings.TrimSuffix(resultText, "```")
	resultText = strings.TrimSpace(resultText)

	var analyzeResp AnalyzeReceiptResponse
	if err := json.Unmarshal([]byte(resultText), &analyzeResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to parse Gemini response: %v", err)})
		return
	}

	c.JSON(http.StatusOK, analyzeResp)
}
