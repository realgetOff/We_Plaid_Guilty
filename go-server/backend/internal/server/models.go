package server

import (
	"github.com/gin-gonic/gin"

	"github.com/realgetOff/We_Plaid_Guilty/internal/db"
	"github.com/realgetOff/We_Plaid_Guilty/internal/gamemanager"
	"github.com/realgetOff/We_Plaid_Guilty/internal/handler"
)

type ServerVarsStruct struct {
	globalHub *gamemanager.Hub
	ClientHub *handler.ClientHub
	router *gin.Engine
	db *db.DBSafe
}