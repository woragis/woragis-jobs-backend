# Job Application Recommendation System - Architecture & Implementation Plan

## Overview

This document outlines the architecture and implementation plan for a recommendation system that provides personalized job application recommendations based on data analysis and machine learning.

## Goals

1. Provide personalized recommendations for job applications
2. Track company-level metrics and analytics
3. Implement company and location deduplication
4. Create an analytics dashboard
5. Integrate recommendations into the jobs UI
6. Build a scalable ML service using Python

## Architecture Decision

### Service Architecture: Separate ML/Recommendation Service

**Structure:**
```
jobs-service/ (Go)
  ├── jobapplications/ (existing)
  └── companies/ (new domain)

ml-recommendation-service/ (Python)
  ├── analytics/
  ├── recommendations/
  ├── models/ (ML training/inference)
  └── company_deduplication/
```

### Communication Pattern

- **Kafka**: Jobs service publishes events, ML service consumes them
- **REST API**: Frontend calls ML service for recommendations and analytics
- **Hybrid Recommendation Retrieval**: Real-time for user's applications, cached for company-level metrics

## Domain Structure

### 1. Company Domain (New - in Jobs Service)

**Purpose:** Store company-level data and metrics

**Entity:**
```go
type Company struct {
    ID              uuid.UUID
    Name            string          // Normalized name
    NormalizedName  string          // For deduplication
    Size            string          // "100-200", "500-1000", etc.
    Industry        string          // Optional, for future
    Location        string          // Normalized location
    NormalizedLocation string       // For deduplication (handles "Remote", "remote", etc.)
    
    // Aggregated metrics (updated by ML service)
    AvgResponseTime        *int      // Days
    AvgTimeToInterview    *int      // Days
    AvgInterviewCount     *float64
    SuccessRate           *float64  // Accepted/Total ratio
    AvgSalaryMin          *int
    AvgSalaryMax          *int
    
    CreatedAt      time.Time
    UpdatedAt      time.Time
}
```

**Features:**
- Company deduplication (normalized name matching)
- Location deduplication (normalizes "Remote", "remote", "REMOTE", etc.)
- Aggregated metrics from all user applications

**Relationship:**
- JobApplication has `company_id` (FK to Company)
- Company created on-the-fly when application is created (or matched via deduplication)

### 2. JobApplication Updates

**New Fields:**
```go
type JobApplication struct {
    // ... existing fields ...
    CompanySize    string    // "100-200", "500-1000", etc.
    CompanyID      *uuid.UUID // FK to Company (nullable for migration)
}
```

### 3. ML Recommendation Service (Python)

**Domains:**
- **Analytics Domain**: Calculate and store metrics
- **Recommendations Domain**: Generate and cache recommendations
- **Company Deduplication**: Match and normalize company names
- **Location Deduplication**: Normalize location strings
- **Models Domain**: ML model training and inference

## Data Flow

### Event-Driven Architecture

**Kafka Topics:**
- `job-applications.created`
- `job-applications.updated`
- `job-applications.status-changed`
- `interviews.added`
- `responses.received`
- `applications.deleted`

**Event Payload Example:**
```json
{
  "event_type": "application.created",
  "application_id": "uuid",
  "user_id": "uuid",
  "company_name": "Company X",
  "company_size": "100-200",
  "location": "Remote",
  "salary_min": 100000,
  "salary_max": 150000,
  "status": "pending",
  "timestamp": "2026-01-06T..."
}
```

### Flow Diagram

```
User creates/updates application
  ↓
Jobs Service (Go)
  ↓
1. Save to DB
2. Attempt company deduplication/match
3. Create/update Company record
4. Publish event to Kafka
  ↓
Kafka
  ↓
ML Service (Python - Kafka Consumer)
  ↓
1. Consume event
2. Update company metrics (aggregate)
3. Update user metrics
4. Recalculate recommendations for user
5. Store in ML service DB
6. Cache in Redis (optional)
  ↓
Frontend requests recommendations
  ↓
ML Service (REST API)
  ↓
1. Check cache (company metrics)
2. Calculate real-time (user applications)
3. Combine and return recommendations
```

## Recommendation System Design

### Scoring Algorithm (Phase 1-2)

**Opportunity Score Calculation:**
```python
def calculate_opportunity_score(application, company_metrics, user_metrics):
    score = 0
    
    # Response speed (0-25 points)
    if company_metrics.avg_response_time:
        if application.response_time < company_metrics.avg_response_time:
            score += 25
        else:
            score += max(0, 25 - (application.response_time - company_metrics.avg_response_time) * 2)
    
    # Salary range (0-25 points)
    if application.salary_max:
        score += min(25, (application.salary_max / 200000) * 25)
    
    # Interest level (0-20 points)
    interest_scores = {"very-high": 20, "high": 15, "medium": 10, "low": 5}
    score += interest_scores.get(application.interest_level, 0)
    
    # Interview progression (0-20 points)
    score += min(20, application.interview_count * 5)
    
    # Time sensitivity (0-10 points)
    if application.deadline:
        days_until = (application.deadline - now).days
        if 0 <= days_until <= 7:
            score += 10
        elif 8 <= days_until <= 14:
            score += 5
    
    return min(100, score)
```

**Tier Classification:**
- **Tier S (90-100)**: Hot opportunities - highest priority
- **Tier A (70-89)**: Strong candidates - high priority
- **Tier B (50-69)**: Waiting - medium priority
- **Tier C (0-49)**: At risk / Low priority

### Recommendation Types

1. **"Hot Opportunities"**: Recent activity, high score, deadlines approaching
2. **"Fast Movers"**: Companies with fastest response times
3. **"High Value"**: Best salary opportunities
4. **"Needs Attention"**: Applications requiring follow-up
5. **"Similar Opportunities"**: Based on successful applications
6. **"At Risk"**: No response for extended period

### Explainability

Each recommendation includes:
- **Score breakdown**: Why this score
- **Key factors**: "Fast response time", "High salary", etc.
- **Comparisons**: "Faster than average", "Above your average salary"
- **Action items**: "Follow up in 3 days", "Deadline approaching"

## Company Deduplication Strategy

### Normalization Rules

**Company Name:**
- Convert to lowercase
- Remove common suffixes: "Inc.", "LLC", "Ltd.", "Corp.", "Corporation"
- Remove special characters: ".", ",", "-", "&" → "and"
- Remove extra whitespace
- Handle common variations: "Google" = "Google Inc." = "Google LLC"

**Location:**
- Normalize "Remote" variations: "Remote", "remote", "REMOTE", "Work from Home", "WFH" → "Remote"
- Normalize city names: "San Francisco" = "SF" = "San Fran"
- Handle country variations: "United States" = "USA" = "US"

### Matching Algorithm

1. **Exact match** on normalized name
2. **Fuzzy match** (Levenshtein distance) for typos
3. **Manual override** flag for user corrections

## Database Design

### Jobs Service Database (PostgreSQL)

**Tables:**
- `companies` (new)
- `job_applications` (add company_id, company_size fields)

### ML Service Database (PostgreSQL)

**Tables:**
- `company_metrics` - Aggregated company statistics
- `user_metrics` - User-level statistics
- `recommendations` - Cached recommendations per user
- `recommendation_history` - Track what was recommended and user actions
- `ml_models` - Model metadata and versions
- `model_predictions` - Store predictions for analysis

**Schema:**
```sql
-- Company Metrics
CREATE TABLE company_metrics (
    company_id UUID PRIMARY KEY,
    total_applications INT,
    avg_response_time_days INT,
    avg_time_to_interview_days INT,
    avg_interview_count DECIMAL,
    success_rate DECIMAL, -- accepted / total
    avg_salary_min INT,
    avg_salary_max INT,
    last_updated TIMESTAMP
);

-- User Metrics
CREATE TABLE user_metrics (
    user_id UUID PRIMARY KEY,
    total_applications INT,
    avg_response_time_days INT,
    success_rate DECIMAL,
    avg_salary_range_min INT,
    avg_salary_range_max INT,
    preferred_company_sizes TEXT[], -- Array of preferred sizes
    last_updated TIMESTAMP
);

-- Recommendations Cache
CREATE TABLE recommendations (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    application_id UUID NOT NULL,
    opportunity_score INT,
    tier VARCHAR(10), -- S, A, B, C
    recommendation_type VARCHAR(50), -- hot_opportunity, fast_mover, etc.
    explanation JSONB, -- Score breakdown and factors
    created_at TIMESTAMP,
    expires_at TIMESTAMP,
    INDEX idx_user_recommendations (user_id, created_at DESC)
);

-- Recommendation History (for learning)
CREATE TABLE recommendation_history (
    id UUID PRIMARY KEY,
    user_id UUID,
    application_id UUID,
    recommended_at TIMESTAMP,
    user_action VARCHAR(50), -- viewed, applied, ignored, etc.
    action_timestamp TIMESTAMP
);
```

## API Design

### ML Service REST Endpoints

**Recommendations:**
```
GET /api/v1/recommendations/{user_id}
  - Returns: List of recommended applications with scores and explanations
  - Query params: limit, tier, type

GET /api/v1/recommendations/{user_id}/hot-opportunities
  - Returns: Top hot opportunities

GET /api/v1/recommendations/{user_id}/needs-attention
  - Returns: Applications requiring follow-up
```

**Analytics:**
```
GET /api/v1/analytics/{user_id}/overview
  - Returns: User metrics and insights

GET /api/v1/analytics/{user_id}/company/{company_id}
  - Returns: Company-specific metrics

GET /api/v1/analytics/{user_id}/trends
  - Returns: Time-series trends
```

**Company Deduplication:**
```
POST /api/v1/companies/deduplicate
  - Body: { "name": "Company X", "location": "Remote" }
  - Returns: { "company_id": "uuid", "matched": true }

GET /api/v1/companies/{company_id}/metrics
  - Returns: Company metrics
```

## Implementation Phases

### Phase 1: Foundation (Week 1-2)

**Jobs Service:**
- [ ] Create Company domain
- [ ] Add company_id and company_size to JobApplication
- [ ] Implement basic company deduplication (exact match)
- [ ] Add Kafka event publishing
- [ ] Migration scripts

**ML Service Setup:**
- [ ] Initialize Python service structure
- [ ] Set up database schema
- [ ] Kafka consumer setup
- [ ] Basic event processing

### Phase 2: Scoring System (Week 3-4)

**ML Service:**
- [ ] Implement opportunity score calculation
- [ ] Company metrics aggregation
- [ ] User metrics calculation
- [ ] Recommendation generation
- [ ] REST API endpoints
- [ ] Hybrid retrieval (real-time + cache)

**Frontend:**
- [ ] Recommendations UI component
- [ ] Display scores and explanations
- [ ] Tier badges/colors

### Phase 3: Deduplication & Analytics (Week 5-6)

**ML Service:**
- [ ] Advanced company deduplication (fuzzy matching)
- [ ] Location normalization
- [ ] Analytics dashboard API
- [ ] Recommendation history tracking

**Frontend:**
- [ ] Analytics dashboard page
- [ ] Charts and visualizations
- [ ] Company metrics display

### Phase 4: Machine Learning (Week 7-8+)

**ML Service:**
- [ ] Collect training data
- [ ] Implement ML models (success prediction, response time prediction)
- [ ] Model training pipeline
- [ ] A/B testing framework
- [ ] Model versioning

## Technology Stack

### ML Service

**Core:**
- Python 3.11+
- FastAPI (REST API)
- SQLAlchemy (ORM)
- Alembic (migrations)

**ML/Analytics:**
- scikit-learn (ML models)
- pandas (data processing)
- numpy (numerical operations)

**Infrastructure:**
- Kafka-Python (Kafka consumer)
- Redis (caching, optional)
- PostgreSQL (database)

**Utilities:**
- python-Levenshtein (fuzzy matching)
- pydantic (data validation)

## Environment Variables

### ML Service

```bash
# Database
ML_DATABASE_URL=postgresql://user:pass@localhost:5433/ml_service

# Kafka
KAFKA_BOOTSTRAP_SERVERS=localhost:9092
KAFKA_GROUP_ID=ml-recommendation-service

# Redis (optional)
REDIS_URL=redis://localhost:6379/1

# API
ML_SERVICE_PORT=3020
ML_SERVICE_HOST=0.0.0.0

# Features
ENABLE_ML_MODELS=true
RECOMMENDATION_CACHE_TTL=3600
```

## Security Considerations

1. **Authentication**: ML service validates JWT tokens from auth service
2. **Authorization**: Users can only access their own recommendations
3. **Data Privacy**: Aggregate metrics only, no PII in ML service
4. **Rate Limiting**: Prevent abuse of recommendation endpoints

## Monitoring & Observability

1. **Metrics**: Recommendation generation time, cache hit rate, model accuracy
2. **Logging**: Structured logging for all events
3. **Tracing**: Distributed tracing across services
4. **Alerts**: Model performance degradation, high error rates

## Future Enhancements

1. **Collaborative Filtering**: Learn from similar users
2. **External Data**: Company reviews, salary data from external APIs
3. **Real-time Updates**: WebSocket for live recommendation updates
4. **Multi-user Support**: Company-level insights across all users
5. **Advanced ML**: Deep learning models for complex patterns
6. **Recommendation Explanations**: More detailed "why" explanations

## Success Metrics

1. **Recommendation Quality**: Click-through rate on recommendations
2. **User Engagement**: Time spent on recommended applications
3. **Conversion Rate**: Applications from recommendations that lead to interviews
4. **Model Accuracy**: Prediction accuracy for success/response time
5. **Performance**: API response times < 200ms for cached, < 500ms for real-time

## Notes

- Start simple with scoring, add ML gradually
- Focus on explainability - users need to understand why
- Collect data early for future ML training
- Keep recommendations user-specific for privacy
- Both analytics dashboard and UI integration provide value

