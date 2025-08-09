# League Stocks

A real-time financial simulation API built with Go, featuring League of Legends themed market tickers with live price movements implemented by Geometric Brownian Motion.

## Features

- **Real-time Price Simulation**: Continuous price updates using stochastic mathematical models
- **RESTful API**: Clean JSON endpoints for market data and user management
- **Time-Series Database**: MongoDB with optimized time-series collections and TTL indexes

## Technologies

Golang, Echo, Docker, MongoDB

## Market Tickers

| Symbol | Name            | Drift | Volatility | Description    |
| ------ | --------------- | ----- | ---------- | -------------- |
| `MNT`  | Mountain Dragon | 12%   | 35%        | Stable growth  |
| `INF`  | Infernal Dragon | 8%    | 45%        | Moderate risk  |
| `ELD`  | Elder Dragon    | 15%   | 60%        | High growth    |
| `CLD`  | Cloud Dragon    | 25%   | 80%        | Ultra volatile |

## Run with Docker

```bash
# Clone and start
git clone <repository-url>
cd api-project

# Start with Docker
docker-compose up --build -d

#See logs
docker ps
docker logs {image_name}

# API available at http://localhost:8080

```

## API Endpoints

### Logging In

Create an Account
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Aatrox",
    "email": "aatrox@darkin.com",
    "password": "fear"
  }'
```
This will output a token - which will be your auth bearer token.

Log In
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "aatrox@darkin.com", 
    "password": "fear"
  }'
```
This will also give you your auth bearer token. 

### Market Data

```bash
# Get all tickers
GET /api/tickers

# Get latest price for a ticker
GET /api/tickers/{symbol}/price
```

### User Management

```bash
# List all users
GET /api/users

# Get specific user
GET /api/users/{id}

# Create new user
POST /api/users
```
These are protected routes. You will have to use your auth bearer wotken like so: 
```bash
-H "Authorization: Bearer (Token goes here)"
```

## Price Simulation

Uses **Geometric Brownian Motion** model:

```
S(t+1) = S(t) × exp((μ - 0.5σ²)Δt + σ√Δt × Z)
```

Where:

- `S(t)` = Current price
- `μ` = Drift (annual return)
- `σ` = Volatility (annual)
- `Δt` = Time step (1 second)
- `Z` = Random normal variable
