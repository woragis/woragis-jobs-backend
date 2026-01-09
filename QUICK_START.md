# 🚀 Resume Worker Integration - Quick Start Guide

## What Was Accomplished

✅ **Resume Service Updated** to v2.0.0  
✅ **Resume Worker Created** in TypeScript  
✅ **Docker Compose Integration** complete  
✅ **Database Schema** with migrations  
✅ **Comprehensive Documentation** included

---

## 📁 What Was Created

```
workers/resume-worker/
├── src/
│   ├── index.ts                 # Entry point
│   ├── config.ts                # Configuration
│   ├── logger.ts                # Logging
│   ├── database.ts              # PostgreSQL client
│   ├── rabbitmq.ts              # RabbitMQ consumer
│   ├── ai-service-client.ts     # AI Service API
│   ├── resume-service-client.ts # Resume Service API
│   └── job-processor.ts         # Job orchestration
├── dist/                        # Compiled JavaScript ✓
├── Dockerfile                   # Multi-stage build
├── package.json                 # Dependencies
├── tsconfig.json                # TypeScript config
├── migrations.sql               # Database schema
├── .env.sample                  # Configuration template
├── README.md                    # Project README
├── INTEGRATION_GUIDE.md         # Detailed guide
└── test-job.sh                  # Test script
```

---

## 🎯 Quick Start (5 minutes)

### 1️⃣ Navigate to Jobs Directory

```bash
cd backend/jobs
```

### 2️⃣ Update Resume Service Version

Already done! ✓ (docker-compose.yml updated to v2.0.0)

### 3️⃣ Build and Start Services

```bash
docker-compose up -d
```

### 4️⃣ Verify All Services Are Healthy

```bash
docker-compose ps
```

Expected output:

```
NAME                            STATUS
woragis-jobs-database          Up (healthy)
woragis-jobs-redis             Up (healthy)
woragis-jobs-rabbitmq          Up (healthy)
woragis-jobs-app               Up
woragis-jobs-ai-service        Up (healthy)
woragis-jobs-creative-service  Up (healthy)
woragis-jobs-resume-service    Up (healthy)
woragis-jobs-resume-worker     Up              ← NEW!
```

### 5️⃣ Check Logs

```bash
docker logs -f woragis-jobs-resume-worker
```

You should see:

```
Resume Worker initialized successfully
Connected to RabbitMQ
Started consuming from queue
```

---

## 🧪 Testing

### Send a Test Resume Job

```bash
cd workers/resume-worker
bash test-job.sh
```

### Monitor Processing

```bash
docker logs -f woragis-jobs-resume-worker
```

### Query Database

```bash
docker exec woragis-jobs-database psql -U woragis -d jobs_service -c \
  "SELECT id, user_id, status, created_at FROM resume_jobs LIMIT 5;"
```

---

## 📋 Workflow

```
┌─────────────────────────────────┐
│  Job Service publishes request  │
│  to RabbitMQ queue              │
└────────────────┬────────────────┘
                 │
                 ▼
┌─────────────────────────────────┐
│  Resume Worker receives job     │
│  Updates status → processing    │
└────────────────┬────────────────┘
                 │
         ┌───────┴────────┐
         │                │
         ▼                ▼
    ┌────────────┐  ┌──────────────┐
    │AI Service  │  │Resume Service│
    │(generates  │  │(generates    │
    │content)    │  │PDF)          │
    └────┬───────┘  └──────┬───────┘
         │                 │
         └────────┬────────┘
                  │
                  ▼
    ┌─────────────────────────────┐
    │Stores resume in database    │
    │Updates status → completed   │
    │Acknowledges RabbitMQ msg    │
    └─────────────────────────────┘
```

---

## 🔧 Configuration

All settings are in `workers/resume-worker/.env.sample`

Key variables:

- `RABBITMQ_HOST`: Queue server (default: woragis-jobs-rabbitmq)
- `DATABASE_URL`: PostgreSQL connection
- `RESUME_SERVICE_URL`: Resume service endpoint
- `AI_SERVICE_URL`: AI service endpoint
- `LOG_LEVEL`: Logging verbosity (debug|info|warn|error)

---

## 📊 Database Operations

### View Resume Jobs

```bash
docker exec woragis-jobs-database psql -U woragis -d jobs_service -c \
  "SELECT id, user_id, status FROM resume_jobs;"
```

### View Generated Resumes

```bash
docker exec woragis-jobs-database psql -U woragis -d jobs_service -c \
  "SELECT id, job_id, file_path FROM resumes;"
```

### Apply Migrations (if needed)

```bash
docker exec woragis-jobs-database psql -U woragis -d jobs_service < \
  workers/resume-worker/migrations.sql
```

---

## 🐛 Troubleshooting

### Worker Won't Start

```bash
# Check logs
docker logs woragis-jobs-resume-worker

# Check dependencies are healthy
docker-compose ps

# Check RabbitMQ connection
docker logs woragis-jobs-rabbitmq | tail -20
```

### Resume Service Connection Error

```bash
# Check resume-service is running
docker logs woragis-jobs-resume-service

# Test connectivity from worker
docker exec woragis-jobs-resume-worker curl -v http://woragis-jobs-resume-service:8080/healthz
```

### Database Connection Issues

```bash
# Check database is running
docker logs woragis-jobs-database

# Test connection
docker exec woragis-jobs-database psql -U woragis -d jobs_service -c "SELECT 1;"
```

---

## 📚 Documentation

- **[RESUME_WORKER_INTEGRATION.md](./RESUME_WORKER_INTEGRATION.md)** - Complete integration guide
- **[workers/resume-worker/README.md](./workers/resume-worker/README.md)** - Project overview
- **[workers/resume-worker/INTEGRATION_GUIDE.md](./workers/resume-worker/INTEGRATION_GUIDE.md)** - Detailed technical guide
- **[workers/resume-worker/migrations.sql](./workers/resume-worker/migrations.sql)** - Database schema

---

## 🔄 Integration Overview

### Services Communication

```
┌─────────────┐
│ Jobs API    │──┐
└─────────────┘  │
                 └──▶ RabbitMQ ──▶ Resume Worker
                                       ├──▶ Resume Service ──▶ PDF
                                       ├──▶ AI Service ──▶ Content
                                       └──▶ PostgreSQL ──▶ Metadata
```

### Data Flow

1. **Job Submission**: Jobs API → RabbitMQ
2. **Job Processing**: Resume Worker consumes from queue
3. **Content Generation**: Resume Worker → AI Service
4. **PDF Generation**: Resume Worker → Resume Service
5. **Persistence**: Generated resume → PostgreSQL
6. **Completion**: Status update → RabbitMQ acknowledgment

---

## ✨ Key Features

- ✅ **Production-Ready**: Error handling, logging, health checks
- ✅ **Scalable**: Multi-worker capable, load-balanced via RabbitMQ
- ✅ **Type-Safe**: Full TypeScript with strict mode
- ✅ **Resilient**: Retry logic, graceful shutdown, connection pooling
- ✅ **Observable**: Structured logging, correlation IDs
- ✅ **Performant**: Job queuing, connection pooling, concurrent processing

---

## 📦 Tech Stack

- **Runtime**: Node.js 18+ LTS
- **Language**: TypeScript 5.3+
- **Queue**: RabbitMQ 3.13+
- **Database**: PostgreSQL 15+
- **API Client**: Axios
- **Logging**: Pino
- **Build**: Docker (multi-stage)

---

## 🚢 Next Steps

1. **Verify** services are healthy: `docker-compose ps`
2. **Check** worker logs: `docker logs -f woragis-jobs-resume-worker`
3. **Send** test job: `bash workers/resume-worker/test-job.sh`
4. **Monitor** job status: Query PostgreSQL
5. **Scale** if needed: Run additional worker containers

---

## 💡 Pro Tips

### Scaling Workers

To process more jobs concurrently, scale the worker service:

```bash
docker-compose up -d --scale woragis-jobs-resume-worker=3
```

### Debugging

Enable debug logging:

```bash
docker-compose exec woragis-jobs-resume-worker bash
# Inside container:
export LOG_LEVEL=debug
npm start
```

### Monitoring Queue Depth

```bash
docker exec woragis-jobs-rabbitmq rabbitmqctl list_queues name messages
```

### Performance Tuning

Edit environment variables in docker-compose.yml:

- `WORKER_CONCURRENCY`: Jobs processed simultaneously
- `DATABASE_POOL_SIZE`: DB connections
- `RABBITMQ_PREFETCH_COUNT`: Jobs fetched from queue

---

## 📞 Support

For issues or questions:

1. Check logs: `docker logs <service-name>`
2. Review documentation in `workers/resume-worker/`
3. Check database for job status
4. Verify all service health checks pass

---

## ✅ Checklist

- [x] Resume service updated to v2.0.0
- [x] Resume worker created in TypeScript
- [x] Docker compose integration
- [x] Database schema with migrations
- [x] Configuration templates
- [x] Comprehensive documentation
- [x] Test scripts included
- [x] TypeScript compilation verified
- [x] Error handling implemented
- [x] Graceful shutdown handling

**Status: ✅ READY FOR DEPLOYMENT**

---

**Created**: January 8, 2026  
**Version**: 1.0.0  
**Last Updated**: Complete Integration
