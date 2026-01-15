package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type ExportData struct {
	Users    []map[string]interface{} `json:"users"`
	Subjects []map[string]interface{} `json:"subjects"`
	Topics   []map[string]interface{} `json:"topics"`
	Links    []map[string]interface{} `json:"links"`
	Tags     []map[string]interface{} `json:"tags"`
	LinkTags []map[string]interface{} `json:"link_tags"`
}

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "../linky.db.backup"
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	data := ExportData{}

	// Export users
	fmt.Println("Exporting users...")
	rows, err := db.Query("SELECT id, email, password_hash FROM users")
	if err != nil {
		log.Printf("Error querying users: %v", err)
	} else {
		for rows.Next() {
			var id int
			var email, hash string
			rows.Scan(&id, &email, &hash)
			data.Users = append(data.Users, map[string]interface{}{
				"id": id, "email": email, "password_hash": hash,
			})
		}
		rows.Close()
	}

	// Export subjects
	fmt.Println("Exporting subjects...")
	rows, err = db.Query("SELECT id, user_id, name FROM subjects")
	if err != nil {
		log.Printf("Error querying subjects: %v", err)
	} else {
		for rows.Next() {
			var id, userId int
			var name string
			rows.Scan(&id, &userId, &name)
			data.Subjects = append(data.Subjects, map[string]interface{}{
				"id": id, "user_id": userId, "name": name,
			})
		}
		rows.Close()
	}

	// Export topics
	fmt.Println("Exporting topics...")
	rows, err = db.Query("SELECT id, subject_id, name FROM topics")
	if err != nil {
		log.Printf("Error querying topics: %v", err)
	} else {
		for rows.Next() {
			var id, subjectId int
			var name string
			rows.Scan(&id, &subjectId, &name)
			data.Topics = append(data.Topics, map[string]interface{}{
				"id": id, "subject_id": subjectId, "name": name,
			})
		}
		rows.Close()
	}

	// Export links
	fmt.Println("Exporting links...")
	rows, err = db.Query("SELECT id, topic_id, title, url FROM links")
	if err != nil {
		log.Printf("Error querying links: %v", err)
	} else {
		for rows.Next() {
			var id, topicId int
			var title, url string
			rows.Scan(&id, &topicId, &title, &url)
			data.Links = append(data.Links, map[string]interface{}{
				"id": id, "topic_id": topicId, "title": title, "url": url,
			})
		}
		rows.Close()
	}

	// Export tags
	fmt.Println("Exporting tags...")
	rows, err = db.Query("SELECT id, user_id, name FROM tags")
	if err != nil {
		log.Printf("Error querying tags (table may not exist): %v", err)
	} else {
		for rows.Next() {
			var id, userId int
			var name string
			rows.Scan(&id, &userId, &name)
			data.Tags = append(data.Tags, map[string]interface{}{
				"id": id, "user_id": userId, "name": name,
			})
		}
		rows.Close()
	}

	// Export link_tags
	fmt.Println("Exporting link_tags...")
	rows, err = db.Query("SELECT link_id, tag_id FROM link_tags")
	if err != nil {
		log.Printf("Error querying link_tags (table may not exist): %v", err)
	} else {
		for rows.Next() {
			var linkId, tagId int
			rows.Scan(&linkId, &tagId)
			data.LinkTags = append(data.LinkTags, map[string]interface{}{
				"link_id": linkId, "tag_id": tagId,
			})
		}
		rows.Close()
	}

	jsonData, _ := json.MarshalIndent(data, "", "  ")
	err = os.WriteFile("linky-export.json", jsonData, 0644)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n✅ Export completed successfully!\n")
	fmt.Printf("   Exported to: linky-export.json\n\n")
	fmt.Printf("Summary:\n")
	fmt.Printf("  - %d users\n", len(data.Users))
	fmt.Printf("  - %d subjects\n", len(data.Subjects))
	fmt.Printf("  - %d topics\n", len(data.Topics))
	fmt.Printf("  - %d links\n", len(data.Links))
	fmt.Printf("  - %d tags\n", len(data.Tags))
	fmt.Printf("  - %d link_tags\n", len(data.LinkTags))
}
