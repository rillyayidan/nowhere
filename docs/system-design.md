# NowHere System Design

## Main Layers

NowHere has four main layers:

1. Frontend PWA
2. Backend API
3. Context Aggregation Services
4. AI Decision Engine + Preference Storage

## High-Level Flow

```txt
User opens app
    ↓
Frontend asks for GPS permission
    ↓
Frontend sends location to backend
    ↓
Backend fetches:
- weather
- nearby places
- user preferences
    ↓
Backend builds Gemini prompt
    ↓
Gemini returns one recommendation in JSON
    ↓
Frontend displays decision card
    ↓
User accepts or rejects
    ↓
Backend stores feedback