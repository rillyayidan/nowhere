# NowHere

NowHere is a context-aware recommendation app that helps users decide where to go based on location, weather, time, nearby places, and personal preferences.

## Core Idea

User opens the app, NowHere reads real-time context, then Gemini recommends one place with a clear reason.

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
- Tailwind CSS
- PWA-ready structure

### Backend
- Go
- Fiber
- Gemini API
- Google Places API
- Open-Meteo API
- Firestore

### Deployment
- Frontend: Vercel or Firebase Hosting
- Backend: Google Cloud Run

## Project Structure

```txt
nowhere/
├── frontend/
├── backend/
├── docs/
└── README.md