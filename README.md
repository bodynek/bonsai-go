# Bonsai - URL shortening service

Ideally use Docker Compose v2

```bash
docker compose build
docker compose up -d
```

## Bonsai-cli

Commonly needed operations are following

```bash
docker exec -it bonsaid /bonsai-cli list
docker exec -it bonsaid /bonsai-cli add home http://shima.cz
docker exec -it bonsaid /bonsai-cli get home
docker exec -it bonsaid /bonsai-cli delete home
```