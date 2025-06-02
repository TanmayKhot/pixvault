# Pixvault

PixVault is a secure photo management platform built with Go and styled with Tailwind CSS. It offers private galleries, email authentication, and secure storage for organizing and sharing your digital memories.

## Features

### Authentication & Security
- Email/password authentication with secure session management
- Password reset via email
- Email-based sign-in with token verification
- CSRF protection on all forms
- Secure password hashing with bcrypt

### Gallery & Photo Management
- Create, edit, and delete galleries
- Upload multiple images (JPG, PNG, JPEG, GIF)
- Organize photos into galleries
- Toggle gallery privacy (public/private)
- Delete individual images
- View images in gallery format

## Tech Stack

- **Backend**: Go
- **Frontend**: HTML, Tailwind CSS
- **Database**: PostgreSQL
- **Email**: SMTP
- **Containerization**: Docker
- **Development**: Docker Compose, Adminer (DB management)

## Setup

1. Copy `.env.template` to `.env` and configure:
   - Database settings (PSQL_*)
   - SMTP settings
   - CSRF key
   - Server address

2. Run with Docker:
```bash
# Development
docker compose up

# Production
docker compose -f docker-compose.yml -f docker-compose.production.yml up
```

## Project Structure

```
pixvault/
├── assets/          # Static assets
├── cmd/            # Command-line applications
├── controllers/    # HTTP request handlers
├── migrations/     # Database migrations
├── models/         # Database models
├── templates/      # HTML templates
├── views/          # View logic
└── tailwind/       # Tailwind CSS configuration
```
