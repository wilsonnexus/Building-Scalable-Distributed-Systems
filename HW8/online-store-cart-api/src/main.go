package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	_ "github.com/go-sql-driver/mysql"

	"github.com/google/uuid"
)

type Config struct {
	Port        string
	DBMode      string
	AWSRegion   string
	MySQLDSN    string
	DynamoTable string
}

type ShoppingCart struct {
	CartID         string     `json:"cart_id" dynamodbav:"cart_id"`
	CustomerID     string     `json:"customer_id" dynamodbav:"customer_id"`
	CustomerEmail  string     `json:"customer_email" dynamodbav:"customer_email"`
	Status         string     `json:"status" dynamodbav:"status"`
	CreatedAt      string     `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt      string     `json:"updated_at" dynamodbav:"updated_at"`
	Items          []CartItem `json:"items" dynamodbav:"items"`
}

type CartItem struct {
	ProductID string  `json:"product_id" dynamodbav:"product_id"`
	Name      string  `json:"name" dynamodbav:"name"`
	Quantity  int     `json:"quantity" dynamodbav:"quantity"`
	UnitPrice float64 `json:"unit_price" dynamodbav:"unit_price"`
}

type CreateCartRequest struct {
	CustomerID    string `json:"customer_id"`
	CustomerEmail string `json:"customer_email"`
}

type AddItemRequest struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type Server struct {
	cfg   Config
	mysql *sql.DB
	ddb   *dynamodb.Client
}

func main() {
	cfg := Config{
		Port:        getenv("PORT", "8080"),
		DBMode:      getenv("DB_MODE", "mysql"), // mysql or dynamodb
		AWSRegion:   getenv("AWS_REGION", "us-east-1"),
		MySQLDSN:    os.Getenv("MYSQL_DSN"),
		DynamoTable: getenv("DYNAMODB_TABLE", "shopping_carts"),
	}

	s := &Server{cfg: cfg}

	if cfg.DBMode == "mysql" {
		db, err := sql.Open("mysql", cfg.MySQLDSN)
		if err != nil {
			log.Fatalf("mysql open failed: %v", err)
		}

		db.SetMaxOpenConns(20)
		db.SetMaxIdleConns(10)
		db.SetConnMaxLifetime(5 * time.Minute)
		db.SetConnMaxIdleTime(2 * time.Minute)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			log.Fatalf("mysql ping failed: %v", err)
		}

		if err := initMySQLSchema(db); err != nil {
			log.Fatalf("init schema failed: %v", err)
		}

		s.mysql = db
		log.Println("running in MySQL mode")
	} else {
		cfgAWS, err := awsconfig.LoadDefaultConfig(context.Background(), awsconfig.WithRegion(cfg.AWSRegion))
		if err != nil {
			log.Fatalf("aws config failed: %v", err)
		}
		s.ddb = dynamodb.NewFromConfig(cfgAWS)
		log.Println("running in DynamoDB mode")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/shopping-carts", s.handleShoppingCarts)
	mux.HandleFunc("/shopping-carts/", s.handleShoppingCartByPath)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: recoverMiddleware(loggingMiddleware(mux)),
	}

	log.Printf("listening on :%s", cfg.Port)
	log.Fatal(server.ListenAndServe())
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"db_mode": s.cfg.DBMode,
	})
}

func (s *Server) handleShoppingCarts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.handleCreateCart(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
	}
}

func (s *Server) handleShoppingCartByPath(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/shopping-carts/")
	parts := strings.Split(strings.Trim(path, "/"), "/")

	if len(parts) == 1 && r.Method == http.MethodGet {
		s.handleGetCart(w, r, parts[0])
		return
	}

	if len(parts) == 2 && parts[1] == "items" && r.Method == http.MethodPost {
		s.handleAddItem(w, r, parts[0])
		return
	}

	writeError(w, http.StatusNotFound, "NOT_FOUND", "route not found")
}

func (s *Server) handleCreateCart(w http.ResponseWriter, r *http.Request) {
	var req CreateCartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "invalid json")
		return
	}

	if strings.TrimSpace(req.CustomerID) == "" || strings.TrimSpace(req.CustomerEmail) == "" {
		writeError(w, http.StatusBadRequest, "INVALID_INPUT", "customer_id and customer_email are required")
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)
	cart := ShoppingCart{
		CartID:        uuid.NewString(),
		CustomerID:    req.CustomerID,
		CustomerEmail: req.CustomerEmail,
		Status:        "active",
		CreatedAt:     now,
		UpdatedAt:     now,
		Items:         []CartItem{},
	}

	var err error
	if s.cfg.DBMode == "mysql" {
		err = s.createCartMySQL(r.Context(), cart)
	} else {
		err = s.createCartDynamo(r.Context(), cart)
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "CREATE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, cart)
}

func (s *Server) handleGetCart(w http.ResponseWriter, r *http.Request, cartID string) {
	if strings.TrimSpace(cartID) == "" {
		writeError(w, http.StatusBadRequest, "INVALID_INPUT", "cart id required")
		return
	}

	var (
		cart ShoppingCart
		err  error
	)

	if s.cfg.DBMode == "mysql" {
		cart, err = s.getCartMySQL(r.Context(), cartID)
	} else {
		cart, err = s.getCartDynamo(r.Context(), cartID)
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || err.Error() == "cart not found" {
			writeError(w, http.StatusNotFound, "CART_NOT_FOUND", "cart not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "GET_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, cart)
}

func (s *Server) handleAddItem(w http.ResponseWriter, r *http.Request, cartID string) {
	var req AddItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "invalid json")
		return
	}

	if strings.TrimSpace(cartID) == "" || strings.TrimSpace(req.ProductID) == "" || strings.TrimSpace(req.Name) == "" || req.Quantity <= 0 || req.UnitPrice < 0 {
		writeError(w, http.StatusBadRequest, "INVALID_INPUT", "cart_id, product_id, name, quantity > 0, and unit_price >= 0 are required")
		return
	}

	item := CartItem{
		ProductID: req.ProductID,
		Name:      req.Name,
		Quantity:  req.Quantity,
		UnitPrice: req.UnitPrice,
	}

	var (
		cart ShoppingCart
		err  error
	)

	if s.cfg.DBMode == "mysql" {
		cart, err = s.addItemMySQL(r.Context(), cartID, item)
	} else {
		cart, err = s.addItemDynamo(r.Context(), cartID, item)
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || err.Error() == "cart not found" {
			writeError(w, http.StatusNotFound, "CART_NOT_FOUND", "cart not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "ADD_ITEM_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, cart)
}

func initMySQLSchema(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS shopping_carts (
			cart_id VARCHAR(64) PRIMARY KEY,
			customer_id VARCHAR(64) NOT NULL,
			customer_email VARCHAR(255) NOT NULL,
			status VARCHAR(32) NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			INDEX idx_customer_id (customer_id),
			INDEX idx_created_at (created_at)
		)`,
		`CREATE TABLE IF NOT EXISTS shopping_cart_items (
			cart_id VARCHAR(64) NOT NULL,
			product_id VARCHAR(64) NOT NULL,
			name VARCHAR(255) NOT NULL,
			quantity INT NOT NULL,
			unit_price DECIMAL(10,2) NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			PRIMARY KEY (cart_id, product_id),
			CONSTRAINT fk_cart_items_cart FOREIGN KEY (cart_id) REFERENCES shopping_carts(cart_id) ON DELETE CASCADE
		)`,
	}

	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) createCartMySQL(ctx context.Context, cart ShoppingCart) error {
	tx, err := s.mysql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		`INSERT INTO shopping_carts (cart_id, customer_id, customer_email, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		cart.CartID, cart.CustomerID, cart.CustomerEmail, cart.Status, parseRFC3339(cart.CreatedAt), parseRFC3339(cart.UpdatedAt),
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Server) getCartMySQL(ctx context.Context, cartID string) (ShoppingCart, error) {
	var cart ShoppingCart
	var createdAt, updatedAt time.Time

	err := s.mysql.QueryRowContext(ctx,
		`SELECT cart_id, customer_id, customer_email, status, created_at, updated_at
		 FROM shopping_carts
		 WHERE cart_id = ?`, cartID,
	).Scan(&cart.CartID, &cart.CustomerID, &cart.CustomerEmail, &cart.Status, &createdAt, &updatedAt)
	if err != nil {
		return ShoppingCart{}, err
	}

	rows, err := s.mysql.QueryContext(ctx,
		`SELECT product_id, name, quantity, unit_price
		 FROM shopping_cart_items
		 WHERE cart_id = ?
		 ORDER BY product_id`, cartID)
	if err != nil {
		return ShoppingCart{}, err
	}
	defer rows.Close()

	items := []CartItem{}
	for rows.Next() {
		var item CartItem
		if err := rows.Scan(&item.ProductID, &item.Name, &item.Quantity, &item.UnitPrice); err != nil {
			return ShoppingCart{}, err
		}
		items = append(items, item)
	}

	cart.CreatedAt = createdAt.UTC().Format(time.RFC3339)
	cart.UpdatedAt = updatedAt.UTC().Format(time.RFC3339)
	cart.Items = items

	return cart, nil
}

func (s *Server) addItemMySQL(ctx context.Context, cartID string, item CartItem) (ShoppingCart, error) {
	tx, err := s.mysql.BeginTx(ctx, nil)
	if err != nil {
		return ShoppingCart{}, err
	}
	defer tx.Rollback()

	var exists string
	err = tx.QueryRowContext(ctx, `SELECT cart_id FROM shopping_carts WHERE cart_id = ?`, cartID).Scan(&exists)
	if err != nil {
		return ShoppingCart{}, err
	}

	now := time.Now().UTC()
	_, err = tx.ExecContext(ctx,
		`INSERT INTO shopping_cart_items (cart_id, product_id, name, quantity, unit_price, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE
		   name = VALUES(name),
		   quantity = VALUES(quantity),
		   unit_price = VALUES(unit_price),
		   updated_at = VALUES(updated_at)`,
		cartID, item.ProductID, item.Name, item.Quantity, item.UnitPrice, now, now,
	)
	if err != nil {
		return ShoppingCart{}, err
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE shopping_carts SET updated_at = ? WHERE cart_id = ?`,
		now, cartID,
	)
	if err != nil {
		return ShoppingCart{}, err
	}

	if err := tx.Commit(); err != nil {
		return ShoppingCart{}, err
	}

	return s.getCartMySQL(ctx, cartID)
}

func (s *Server) createCartDynamo(ctx context.Context, cart ShoppingCart) error {
	av, err := attributevalue.MarshalMap(cart)
	if err != nil {
		return err
	}

	_, err = s.ddb.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.cfg.DynamoTable),
		Item:      av,
	})
	return err
}

func (s *Server) getCartDynamo(ctx context.Context, cartID string) (ShoppingCart, error) {
	out, err := s.ddb.GetItem(ctx, &dynamodb.GetItemInput{
		TableName:      aws.String(s.cfg.DynamoTable),
		Key:            map[string]ddbtypes.AttributeValue{"cart_id": &ddbtypes.AttributeValueMemberS{Value: cartID}},
		ConsistentRead: aws.Bool(false),
	})
	if err != nil {
		return ShoppingCart{}, err
	}
	if out.Item == nil {
		return ShoppingCart{}, errors.New("cart not found")
	}

	var cart ShoppingCart
	if err := attributevalue.UnmarshalMap(out.Item, &cart); err != nil {
		return ShoppingCart{}, err
	}
	if cart.Items == nil {
		cart.Items = []CartItem{}
	}
	return cart, nil
}

func (s *Server) addItemDynamo(ctx context.Context, cartID string, item CartItem) (ShoppingCart, error) {
	cart, err := s.getCartDynamo(ctx, cartID)
	if err != nil {
		return ShoppingCart{}, err
	}

	found := false
	for i := range cart.Items {
		if cart.Items[i].ProductID == item.ProductID {
			cart.Items[i] = item
			found = true
			break
		}
	}
	if !found {
		cart.Items = append(cart.Items, item)
	}
	cart.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	av, err := attributevalue.MarshalMap(cart)
	if err != nil {
		return ShoppingCart{}, err
	}

	_, err = s.ddb.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.cfg.DynamoTable),
		Item:      av,
	})
	if err != nil {
		return ShoppingCart{}, err
	}

	return cart, nil
}

func parseRFC3339(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Now().UTC()
	}
	return t
}

func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("%v", rec))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s in %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, ErrorResponse{
		Error:   code,
		Message: message,
	})
}

func getenv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
