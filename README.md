
## Raisonnement 

La première étape de l'exercice consiste à récupérer les données des 100 derniers dépôts publics. La seconde étape consiste à filtrer et structurer ces données le plus rapidement possible (en multi-threading via les goroutines) au moment où une requête est faite sur notre endpoint. Il devra également comporter des query params pour filtrer les données (langage, licence, etc.) 

```
Exemple de filtrage : /api/endpoint?lang=go&license=MIT
```

Il est important de penser performance et scalabilité dès le départ. Si nous avons 1 million de requêtes/s sur cet endpoint, nous ne pouvons pas nous permettre de faire 1 million de requêtes/s vers l'API GitHub. Il est donc primordial de mettre en place un système de cache pour servir les mêmes données en notre possession à l'ensemble des clients.

L'enjeu majeur est de servir les données les plus "fraîches" possibles provenant de GitHub aux clients sans pour autant atteindre le "rate limit" de l'API GitHub qui sont les suivantes :

- Primary rate limit for unauthenticated users : 60 requests per hour ou 0,0167 requêtes par seconde ou 1 requête par minute
- Primary rate limit for GitHub App installations : 5000 requests per hour ou 1,3889 requêtes par seconde

Pour notre exercice, dans un souci de simplification, nous allons rester sur le rate limit pour les utilisateurs non-authentifiés. 

Donc pour avoir les données les plus fraîches possibles, nous pouvons interroger l'API de GitHub toutes les minutes je pourrais mettre en place un scheduler qui fetch toutes les 1 minutes mais je ne pense pas que cela vaille la peine de générer du traffic inutile. Je préfère que la 1ère requête toutes les 1 minutes soit un peu plus lente à être disponible (Il s'agit plus d'une prise de parti sujet à discussion bien entendu 🙂)

Note :
Il semblerait que GitHub implémente son propre système de cache côté serveur. Pour l'API de recherche, la durée de ce cache semble être d'environ 30 secondes à 1 minute, bien que ce ne soit pas officiellement documenté.

## Plan d'action / Pseudo code

1. Lorsqu'une requête arrive sur notre endpoint :
   - Vérifier si des données sont présentes dans le cache.
   - Si le cache est vide ou si les données sont périmées (plus de 1 minute), déclencher une mise à jour du cache.
   - Si une mise à jour du cache est en cours, attendre qu'elle soit terminée (avec un timeout raisonnable).
   - Retourner les données du cache (mises à jour ou non) au client.

2. Mise à jour du cache :
   - Vérifier la dernière fois que l'API GitHub a été interrogée.
   - Si moins d'une minute s'est écoulée depuis la dernière requête, attendre le temps restant.
   - Interroger l'API GitHub pour obtenir les 100 derniers dépôts publics.
   - Traiter et structurer les données reçues en utilisant des goroutines pour optimiser les performances.
   - Mettre à jour le cache avec les nouvelles données.

3. Gestion du cache :
   - Utiliser une structure de données en mémoire pour stocker les informations des dépôts.
   - Implémenter un mécanisme de verrouillage (mutex) pour éviter les problèmes de concurrence lors des mises à jour.

4. Optimisations et features possiblees :
   - Stocker les données filtrées par langage, licence, etc. pour optimiser les requêtes fréquentes.
   - S'authentifier auprès de GitHub pour augmenter le rate limit de l'API GitHub (bien que leur cache interne semble être la réelle limite)
   - Implémenter un middleware de rate limiting sur nos endpoints pour se protéger contre les abus.
   - Utiliser un cache distribué (comme Redis) si l'application doit être déployée sur plusieurs instances (scaling horizontal).
   - Conserver les données en base de données (ajout de timestamp pour créer un historique chronologique)
   - Proposer d'autres endpoints (exemple: des stats type "language le plus utilisé sur le mois dernier")

## Structure du projet

```
.
├── Dockerfile
├── docker-compose.yml
├── README.md
├── go.mod
├── go.sum
├── cmd
│   └── api
│       └── main.go
├── internal
│   ├── api
│   │   ├── handlers
│   │   │   └── repository_handler.go
│   │   ├── routes.go
│   │   └── server.go
│   ├── cache
│   │   └── cache.go
│   ├── config
│   │   └── config.go
│   ├── github
│   │   └── client.go
│   ├── models
│   │   └── repository.go
│   └── services
│       └── repository_service.go
├── pkg
│   └── utils
│       └── helpers.go
└── tests
    └── integration_test.go
```

## Explication de la structure du projet

### Dossier `cmd`

- `cmd/api/main.go`: Point d'entrée de l'application. 

### Dossier `internal`

Contient le code spécifique à l'application, non destiné à être importé par d'autres projets.

- `api/`: Logique liée à l'API HTTP.
  - `handlers/repository_handler.go`: Gestion des requêtes HTTP pour les dépôts.
  - `routes.go`: Définition des routes de l'API.
  - `server.go`: Configuration et lancement du serveur HTTP.
- `cache/cache.go`: Implémentation de la logique de mise en cache.
- `config/config.go`: Gestion de la configuration de l'application.
- `github/client.go`: Client pour interagir avec l'API GitHub.
- `models/repository.go`: Structures de données pour les dépôts.
- `services/repository_service.go`: Logique pour la gestion des dépôts.

### Dossier `pkg`

Contient du code potentiellement réutilisables dans d'autres projets

- `utils/helpers.go`: Fonctions utilitaires potentiellement réutilisables dans d'autres projets.

### Dossier `tests`

- `integration_test.go`: Tests d'intégration pour l'application.


## Structure de réponse de github

On utilise une requête "large"" pour récupérer les 100 derniers dépôts publics et on obtenir le plus de données possible dans la réponse.

```
https://api.github.com/search/repositories?q=is:public&sort=created&order=desc&per_page=100
```

### Structure de la réponse github


## Structure de l'affichage des données attendu 

### Json output

```json
{
  "repositories": [
    {
      "full_name": "FreeCodeCamp/FreeCodeCamp",
      "owner": "FreeCodeCamp",
      "repository": "FreeCodeCamp",
      "languages": {
        "javascript": {
          "bytes": 123456
        }
      },
      "license": "BSD
    },
    // ...
  ]
}
```

### Go Type Struct

```go
type Response struct {
    Repositories []Repository `json:"repositories"`
}

type Repository struct {
    FullName   string              `json:"full_name"`
    Owner      string              `json:"owner"`
    Repository string              `json:"repository"`
    Languages  map[string]Language `json:"languages"`
}

type Language struct {
    Bytes int `json:"bytes"`
}
```

## Execution

```
docker compose up
```

Application will be then running on port `5000`

## Test

```
$ curl localhost:5000/ping
{ "status": "pong" }
```
