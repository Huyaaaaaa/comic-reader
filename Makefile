.PHONY: backend frontend

backend:
	cd backend && go run ./cmd/server

frontend:
	cd frontend && npm run dev
