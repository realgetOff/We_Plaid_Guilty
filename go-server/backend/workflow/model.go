// package workflow
// 
// import (
	// "github.com/gorilla/websocket"
	// "github.com/jackc/pgx/v5/pgxpool"
	// "main.go/gamemanager"
// )
// 
// type WSContext struct {
	// CurrUsrID *string
	// CurrUsrName *string
	// Db *pgxpool.Pool
	// Conn *websocket.Conn
	// HubAI *gamemanager.AIHub
	// Hub *gamemanager.Hub
	// CurrentRoom *gamemanager.Room
// }
// 
// type HandleFunc func(context *WSContext, msg Message)
// type PipeFunc func(ctx *WSContext, msg Message) bool
// 
// type Dispatcher struct {
	// handlers map[string]HandleFunc
// }
