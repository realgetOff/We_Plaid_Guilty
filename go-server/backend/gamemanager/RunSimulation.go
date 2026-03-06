package gamemanager // Obligatoire car dans le même dossier

import (
	"fmt"
	"time"
)

// On l'appelle TestMain ou ce que tu veux, mais pas "main" 
// si tu veux le lancer via un script ou une commande spécifique.
func RunTestSimulation() {
	// 1. Initialisation : 3 joueurs, rounds de 5s pour le test
	r := NewRoom("DEBUG_ROOM", 5, 3)
	fmt.Println("=== DÉMARRAGE DU TEST GAMEMANAGER ===")

	// 2. Goroutine pour lire les messages envoyés au Front
	go func() {
		for msg := range r.MessageChan {
			fmt.Printf("\n[WS OUT] Player %d reçoit : %+v\n", msg.PlayerID, msg.Data)
		}
	}()

	// 3. Ajout des joueurs
	r.AddPlayer(1, "Alice")
	r.AddPlayer(2, "Bob")
	r.AddPlayer(3, "Charlie")

	// 4. Lancement de la Loop (StartGame lance RunGameLoop en goroutine)
	err := r.StartGame()
	if err != nil {
		fmt.Println("Erreur au start:", err)
		return
	}

	// 5. Simulation du Round 1 : Writing
	time.Sleep(1 * time.Second)
	fmt.Println("\n--- STEP: Round 1 Submissions ---")
	r.SubmiteAction(1, "Un pingouin sur la lune", true)
	r.SubmiteAction(2, "Une frite géante", true)
	r.SubmiteAction(3, "Batman fait les courses", true)

	// 6. Simulation du Round 2 : Drawing
	// On attend que le serveur traite les messages et passe au round suivant
	time.Sleep(1 * time.Second) 
	fmt.Println("\n--- STEP: Round 2 Submissions ---")
	r.SubmiteAction(1, "DESSIN_FRITE", true)    // Alice dessine la frite de Bob
	r.SubmiteAction(2, "DESSIN_BATMAN", true)   // Bob dessine Batman de Charlie
	r.SubmiteAction(3, "DESSIN_PINGOUIN", true) // Charlie dessine le pingouin d'Alice

	// 7. Simulation du Round 3 : Guessing
	time.Sleep(1 * time.Second)
	fmt.Println("\n--- STEP: Round 3 Submissions ---")
	r.SubmiteAction(1, "Je vois une grosse pomme de terre", true)
	r.SubmiteAction(2, "C'est un super-héros au Lidl", true)
	r.SubmiteAction(3, "Un oiseau dans l'espace", true)

	// 8. Fin et Galerie
	time.Sleep(1 * time.Second)
	fmt.Println("\n=== FIN DU TEST. Vérifie les logs de Galerie ci-dessus ===")
}
