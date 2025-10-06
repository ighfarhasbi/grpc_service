package main

// @title Order Client API
// @version 1.0
// @description API for Order Client Service
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email ighfarhasbiash@gmail.com
// @license.name Apache 2.0

// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @schemes http

// @host 172.16.148.101:8083
// @BasePath /
// securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

import (
	"context"
	"log"
	"net/http"
	"os"

	_ "github.com/ighfarhasbi/grpc_service/docs"
	orderpb "github.com/ighfarhasbi/grpc_service/proto/order"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	var port string
	if os.Getenv("APP_ENV") == "local" {
		err := godotenv.Load(".env.client")
		if err != nil {
			log.Fatalf("Error loading .env: %v", err)
		}
		port = os.Getenv("Client_PORT")
	}

	conn, err := grpc.NewClient("order_service:9091", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("client connection error: %v", err)
	}
	defer conn.Close()

	clientConn := orderpb.NewOrderServiceClient(conn)

	e := echo.New()

	// Define routes for swagger
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	e.POST("/order", func(c echo.Context) error {
		return CreateOrder(c, clientConn)
	})
	e.POST("/cancel", func(c echo.Context) error {
		return CancelOrder(c, clientConn)
	})

	// get port from env
	port = os.Getenv("Client_PORT")

	e.Logger.Fatal(e.Start(":" + port))
}

// @Summary Create a new order
// @Description Create a new order with user ID and items
// @Tags orders
// @Accept json
// @Produce json
// @Param order body orderpb.CreateOrderRequest true "Order details"
// @Success 200 {object} orderpb.CreateOrderResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /order [post]
func CreateOrder(c echo.Context, client orderpb.OrderServiceClient) error {
	// bind request body to struct
	var req orderpb.CreateOrderRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request body",
		})
	}

	// call gRPC method
	resp, err := client.CreateOrder(context.Background(), &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to create order: " + err.Error(),
		})
	}

	// return response
	return c.JSON(http.StatusOK, resp)
}

// @Summary Cancel an order
// @Description Cancel an existing order by order ID
// @Tags orders
// @Accept json
// @Produce json
// @Param order body orderpb.CancelOrderRequest true "Order ID"
// @Success 200 {object} orderpb.CancelOrderResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /cancel [post]
func CancelOrder(c echo.Context, client orderpb.OrderServiceClient) error {
	// bind request body to struct
	var req orderpb.CancelOrderRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request body",
		})
	}

	// call gRPC method
	resp, err := client.CancelOrder(context.Background(), &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to cancel order: " + err.Error(),
		})
	}

	// return response
	return c.JSON(http.StatusOK, resp)
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
