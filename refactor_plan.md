1. Créer de nouveaux packages :
   - `pattern` pour les patterns de formation
   - `state` pour la gestion de l'état du jeu
   - `config` pour la configuration

2. Refactoriser `entities/formation.go` :
   - Créer l'interface `Pattern`
   - Implémenter les patterns dans le package `pattern`
   - Modifier `EnemyFormation` pour utiliser l'interface `Pattern`

3. Refactoriser `entities/enemy.go` et `entities/formation.go` :
   - Créer l'interface `Drawable`
   - Faire implémenter `Drawable` par `Enemy` et `EnemyFormation`

4. Refactoriser `game/game.go` :
   - Extraire la logique de gestion d'état dans le package `state`
   - Utiliser les nouvelles interfaces et structures

5. Créer `config/config.go` :
   - Déplacer toutes les constantes et paramètres du jeu dans ce fichier

6. Mettre à jour `graphics/draw.go` :
   - Assurer que toutes les fonctions de dessin sont cohérentes et utilisent les mêmes paramètres

7. Réviser `main.go` :
   - S'assurer qu'il utilise correctement les nouveaux packages et structures

8. Tests unitaires :
   - Ajouter des tests pour chaque package, en se concentrant sur les nouvelles interfaces et structures

9. Revue de code finale :
   - Vérifier la cohérence globale
   - S'assurer qu'il n'y a plus de code dupliqué
   - Vérifier que les noms des fonctions correspondent à leurs actions