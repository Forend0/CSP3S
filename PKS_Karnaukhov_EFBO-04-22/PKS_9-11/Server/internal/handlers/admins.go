package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// GetAdmins возвращает список администраторов
func GetAdmins(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var admins []struct {
			AdminID  int    `db:"admin_id" json:"admin_id"`
			Username string `db:"username" json:"username"`
			Email    string `db:"email" json:"email"`
		}

		err := db.Select(&admins, `
            SELECT admin_id, username, email
            FROM administrators
            WHERE admin_id = 1
        `)

		log.Printf("Found admins: %+v", admins) // Добавляем логирование

		if err != nil {
			log.Printf("Error getting admins: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, admins)
	}
}
