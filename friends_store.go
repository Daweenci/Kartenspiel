package main

import (
	"database/sql"
	"log"
)

func createFriendRequest(playerID, friendID string) bool {
	query := "INSERT INTO friend_requests (sender_id, receiver_id) VALUES (?, ?)"
	_, err := db.Exec(query, playerID, friendID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false
		}
		log.Printf("Error occurred while fetching player ID by name: %v", err)
		return false
	}
	return true
}

func getPendingFriendRequests(playerID string) []FriendRequestDTO {
	query := `
		SELECT fr.sender_id, p.username
		FROM friend_requests fr
		JOIN players p ON fr.sender_id = p.id
		WHERE fr.receiver_id = ? AND fr.status = 'pending'
	`

	rows, err := db.Query(query, playerID)
	if err != nil {
		log.Printf("Error fetching friend requests: %v", err)
		return nil
	}
	defer rows.Close()

	var requests []FriendRequestDTO

	for rows.Next() {
		var fr FriendRequestDTO
		err := rows.Scan(&fr.FriendID, &fr.FriendName)
		if err != nil {
			log.Printf("Error scanning friend request: %v", err)
			continue
		}
		requests = append(requests, fr)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Row iteration error: %v", err)
	}

	return requests
}
