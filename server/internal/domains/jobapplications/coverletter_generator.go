package jobapplications

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"woragis-jobs-service/pkg/aiservice"
)

// AIServiceCoverLetterGenerator implements CoverLetterGenerator using the AI service
type AIServiceCoverLetterGenerator struct {
	client *aiservice.Client
	logger *slog.Logger
}

// NewAIServiceCoverLetterGenerator creates a new AI service cover letter generator
func NewAIServiceCoverLetterGenerator(client *aiservice.Client, logger *slog.Logger) CoverLetterGenerator {
	return &AIServiceCoverLetterGenerator{
		client: client,
		logger: logger,
	}
}

// GenerateCoverLetterWithContext generates a cover letter using the AI service
func (g *AIServiceCoverLetterGenerator) GenerateCoverLetterWithContext(
	ctx context.Context,
	profile UserProfile,
	job JobInfo,
	additionalContext string,
) (string, error) {
	// Build the system prompt for cover letter generation
	systemPrompt := g.buildSystemPrompt()

	// Build the user input with job and profile information
	userInput := g.buildUserInput(profile, job, additionalContext)

	// Call the AI service using the cover_letter agent
	req := aiservice.ChatRequest{
		Agent:  "cover_letter", // Using specialized cover_letter agent
		Input:  userInput,
		System: &systemPrompt,
		// Use default temperature for professional writing (balanced creativity)
		Temperature: func() *float64 { t := 0.7; return &t }(),
		MaxTokens: func() *int { t := 2000; return &t }(), // Cover letters should be concise
	}

	g.logger.Info("generating cover letter",
		"company", job.CompanyName,
		"jobTitle", job.JobTitle,
	)

	resp, err := g.client.Chat(ctx, req)
	if err != nil {
		g.logger.Error("failed to generate cover letter via AI service", "error", err)
		return "", fmt.Errorf("AI service error: %w", err)
	}

	if resp.Output == "" {
		return "", fmt.Errorf("AI service returned empty response")
	}

	g.logger.Info("cover letter generated successfully",
		"length", len(resp.Output),
	)

	return resp.Output, nil
}

// buildSystemPrompt creates the system prompt for cover letter generation
func (g *AIServiceCoverLetterGenerator) buildSystemPrompt() string {
	return `You are an expert career coach and professional writer specializing in crafting compelling, personalized cover letters. 

Your task is to write a professional cover letter that:
1. Is tailored specifically to the job and company
2. Highlights relevant skills and experiences from the candidate's profile
3. Demonstrates genuine interest in the position
4. Is concise (typically 3-4 paragraphs, 250-400 words)
5. Uses professional, confident, and engaging language
6. Avoids generic phrases and clichés
7. Clearly connects the candidate's background to the job requirements

Format the cover letter as a proper business letter with:
- Professional greeting
- Clear, structured paragraphs
- Professional closing

Do not include placeholders or generic text. Make it specific and compelling.`
}

// buildUserInput creates the user input with job and profile information
func (g *AIServiceCoverLetterGenerator) buildUserInput(
	profile UserProfile,
	job JobInfo,
	additionalContext string,
) string {
	var parts []string

	// Job information
	parts = append(parts, "## Job Application Details")
	parts = append(parts, fmt.Sprintf("Company: %s", job.CompanyName))
	parts = append(parts, fmt.Sprintf("Position: %s", job.JobTitle))
	if job.Location != "" {
		parts = append(parts, fmt.Sprintf("Location: %s", job.Location))
	}
	if job.JobDescription != "" {
		parts = append(parts, fmt.Sprintf("Job Description:\n%s", job.JobDescription))
	}
	if len(job.Requirements) > 0 {
		parts = append(parts, fmt.Sprintf("Key Requirements:\n- %s", strings.Join(job.Requirements, "\n- ")))
	}

	// Candidate profile information
	parts = append(parts, "\n## Candidate Profile")
	if len(profile.Skills) > 0 {
		parts = append(parts, fmt.Sprintf("Skills: %s", strings.Join(profile.Skills, ", ")))
	}
	if len(profile.Certifications) > 0 {
		parts = append(parts, fmt.Sprintf("Certifications: %s", strings.Join(profile.Certifications, ", ")))
	}
	if len(profile.Interests) > 0 {
		parts = append(parts, fmt.Sprintf("Interests: %s", strings.Join(profile.Interests, ", ")))
	}
	if len(profile.Projects) > 0 {
		parts = append(parts, fmt.Sprintf("Projects: %d relevant projects", len(profile.Projects)))
	}
	if len(profile.Posts) > 0 {
		parts = append(parts, fmt.Sprintf("Posts/Publications: %d items", len(profile.Posts)))
	}
	if len(profile.TechnicalWritings) > 0 {
		parts = append(parts, fmt.Sprintf("Technical Writings: %d items", len(profile.TechnicalWritings)))
	}

	// Additional context (e.g., from chat messages)
	if additionalContext != "" {
		parts = append(parts, "\n## Additional Context")
		parts = append(parts, additionalContext)
	}

	parts = append(parts, "\n## Instructions")
	parts = append(parts, "Please write a compelling, personalized cover letter for this job application. Make it specific, professional, and engaging.")

	return strings.Join(parts, "\n")
}

