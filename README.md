# NowHere

NowHere is a context-aware recommendation app that helps users decide where to go based on location, weather, time, nearby places, and personal preferences.

## Core Idea

User opens the app, NowHere reads real-time context, then recommends one place with a clear reason.

The app is designed for hackathon/demo use:

1. Get user GPS location
2. Fetch weather context
3. Fetch nearby places
4. Ask Gemini to choose one best recommendation
5. User can accept or reject
6. Feedback is stored to improve future recommendations

## Tech Stack

### Frontend

- React
- Vite
- PWA-ready structure

### Backend

- Go
- Fiber
- Gemini API
- Google Places API
- Open-Meteo API
- Firestore-ready preference boundary

### Deployment

- Frontend: Vercel or Firebase Hosting
- Backend: Google Cloud Run

## Project Structure

```txt
nowhere/
|-- frontend/
|-- backend/
|-- docs/
`-- README.md
```

## Local Development

The MVP runs without API keys by using deterministic demo data. Add keys in
`backend/.env.local` when you want live Vertex AI/Gemini and Google Places
responses.

```bash
# backend
go run ./backend/cmd

# frontend
cd frontend
npm install
npm run dev
```

Open `http://localhost:5173`. The Vite server proxies `/api` requests to
`http://localhost:8080`.

To test the Cloud Run shape locally, build the frontend first and then run the
Go server. The backend will serve `frontend/dist` at `/` and API routes at
`/api`.

```bash
cd frontend && npm run build
cd ..
go run ./backend/cmd
```

## Cloud Run Deployment

The repository includes `cloudbuild.yaml` for a single Cloud Run service that
serves both the React app and the Go API.

Before deploying:

1. Enable Cloud Run, Cloud Build, Artifact Registry, Vertex AI, and Firestore APIs.
2. Create the Artifact Registry repository used by `cloudbuild.yaml`:

```bash
gcloud artifacts repositories create cloud-run \
  --repository-format=docker \
  --location=asia-southeast2
```

3. Grant the Cloud Run runtime service account these roles:

```bash
gcloud projects add-iam-policy-binding PROJECT_ID \
  --member="serviceAccount:PROJECT_NUMBER-compute@developer.gserviceaccount.com" \
  --role="roles/aiplatform.user"

gcloud projects add-iam-policy-binding PROJECT_ID \
  --member="serviceAccount:PROJECT_NUMBER-compute@developer.gserviceaccount.com" \
  --role="roles/datastore.user"
```

4. Deploy:

```bash
gcloud builds submit --config=cloudbuild.yaml
```

5. Make the service public for demos:

```bash
gcloud run services add-iam-policy-binding nowhere \
  --region=asia-southeast2 \
  --member=allUsers \
  --role=roles/run.invoker
```

The default deployment uses Vertex AI with `gemini-2.5-flash-lite`, scales to zero,
and disables demo fallback so production AI/auth failures are visible.

## Build Checks

```bash
go test ./...
cd frontend && npm run build
```
