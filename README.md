# Crême 2.0 🍦

Petit site « depuis quand ? » : des compteurs (titre, phrase, date de départ) qui
tournent seconde par seconde. On peut les créer, modifier, mettre en pause,
reprendre, archiver, et « craquer » (le compteur repart de zéro en gardant
l'historique des tentatives et le record). Chaque compteur peut afficher un
gain qui s'accumule avec le temps (13 € par jour économisés, 3 000 kcal par
jour évitées…) avec des équivalences rigolotes (≈ 9,9 pizzas 🍕), des paliers
fêtés à coups de confettis, une carte PNG à partager, et un mode nuit.
Un seul mot de passe, défini dans le `.env`, avec blocage temporaire d'une IP
après 5 échecs.

- **Front** : Vue 3 + Vite, servi par nginx (qui proxifie `/api` vers le back)
- **Back** : Go (stdlib uniquement), stockage dans un fichier JSON sur un volume Docker
- **Auth** : un mot de passe → jeton HMAC signé, valable 30 jours, sans état côté serveur

## Lancer

Le front est publié sur le réseau Docker externe `web` (celui du reverse proxy :
Traefik, Nginx Proxy Manager, Caddy…). Le back reste sur un réseau interne.

```bash
docker network create web   # si le réseau n'existe pas déjà
cp .env.example .env        # puis change APP_PASSWORD, et PUID/PGID (commande « id »)
docker compose pull && docker compose up -d     # images prêtes à l'emploi (GHCR)
# ou, pour construire soi-même :
docker compose up -d --build
```

## Déploiement

À chaque push sur `main`, la CI GitHub (`.github/workflows/ci.yml`) vérifie le
back (gofmt, vet, tests) et le front (build), puis publie les images
`ghcr.io/hurdmanbegins/creme-backend` et `ghcr.io/hurdmanbegins/creme-frontend`
(tags `latest` et le SHA du commit). Sur le serveur, le projet est un clone git :

```bash
cd /chemin/vers/creme
git pull && docker compose pull && docker compose up -d
```

Côté proxy, la cible est le conteneur `creme-frontend` sur le port 80 (un exemple
de labels Traefik est en commentaire dans `docker-compose.yml`). Pour un accès
direct sans proxy, remplace `expose` par `ports: ["8080:80"]` dans le service
`frontend`.

Les compteurs sont persistés dans `./data/counters.json`. Au démarrage, le
conteneur attribue `data/` à `PUID:PGID` puis lance le serveur sous cet
utilisateur, donc le fichier vous appartient même si Docker a créé le dossier
en root.

## Développement

```bash
# terminal 1 — back sur :8080
cd backend && APP_PASSWORD=chantilly DATA_FILE=./data/counters.json go run .

# terminal 2 — front sur :5173 (proxy /api -> :8080)
cd frontend && npm install && npm run dev
```

## API

Toutes les routes sauf `/api/login` et `/api/health` demandent `Authorization: Bearer <token>`.

| Méthode | Route                        | Rôle                              |
|---------|------------------------------|-----------------------------------|
| POST    | `/api/login`                 | `{ "password" }` → `{ "token" }`  |
| GET     | `/api/me`                    | vérifie le jeton                  |
| GET     | `/api/counters`              | liste + `serverTime`              |
| POST    | `/api/counters`              | crée un compteur                  |
| PUT     | `/api/counters/{id}`         | modifie titre/phrase/emoji/date   |
| DELETE  | `/api/counters/{id}`         | supprime                          |
| POST    | `/api/counters/{id}/pause`   | fige le compteur                  |
| POST    | `/api/counters/{id}/resume`  | reprend (le temps en pause est déduit) |
| POST    | `/api/counters/{id}/reset`   | « j'ai craqué » : archive la tentative, repart de maintenant |
| POST    | `/api/counters/{id}/archive` | range le compteur dans les archives |
| POST    | `/api/counters/{id}/unarchive` | le restaure |

Corps d'un compteur : `{ "title", "phrase", "emoji", "color", "startAt",
"gainAmount", "gainUnit", "gainPeriod", "gainLabel", "equivUnit", "equivValue" }`
(`startAt` en ISO 8601, `color` parmi cream/pink/lavender/mint/sky/peach/lemon,
`gainAmount` ≥ 0 avec 0 = pas de gain, `gainPeriod` parmi hour/day/week/month,
`gainUnit` et `gainLabel` libres, ex. `€` et `économisés`).

Gain affiché = `gainAmount × temps écoulé / période` (figé en pause).
`equivUnit`/`equivValue` : équivalence personnalisée (1 equivUnit = equivValue gainUnit).
Les réponses contiennent aussi `attempts` (tentatives passées), `archived`, `archivedAt`.

Le login renvoie 429 après 5 échecs en 15 min depuis la même IP (première
adresse de `X-Forwarded-For`), pour 15 min.

Temps affiché = `(pausedAt ou maintenant) − startAt − pausedSeconds`.
