package docs

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/swag"
)

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0.0",
	Host:             "localhost:8080",
	BasePath:         "/api/v1",
	Schemes:          []string{"http", "https"},
	Title:            "太上老君 API",
	Description:      "太上老君系统 RESTful API 文档",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

// SetupSwagger 设置 Swagger 文档路由
func SetupSwagger(r *gin.Engine) {
	// Swagger 文档路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API 文档重定向到 Swagger 首页
	r.GET("/docs", func(c *gin.Context) {
		c.Redirect(302, "/swagger/index.html")
	})

	// API 文档首页
	r.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/swagger/index.html")
	})
}

// SwaggerConfig Swagger 配置
type SwaggerConfig struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Host        string   `json:"host"`
	BasePath    string   `json:"base_path"`
	Schemes     []string `json:"schemes"`
	Contact     Contact  `json:"contact"`
	License     License  `json:"license"`
}

// Contact 联系信息
type Contact struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	Email string `json:"email"`
}

// License 许可证信息
type License struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// DefaultSwaggerConfig 默认 Swagger 配置
func DefaultSwaggerConfig() SwaggerConfig {
	return SwaggerConfig{
		Title:       "太上老君 API",
		Description: "太上老君系统 RESTful API 文档，提供完整的用户管理、权限控制、系统配置等功能接口",
		Version:     "1.0.0",
		Host:        "localhost:8080",
		BasePath:    "/api/v1",
		Schemes:     []string{"http", "https"},
		Contact: Contact{
			Name:  "太上老君开发团队",
			URL:   "https://github.com/taishanglaojun",
			Email: "dev@taishanglaojun.com",
		},
		License: License{
			Name: "MIT",
			URL:  "https://opensource.org/licenses/MIT",
		},
	}
}

// UpdateSwaggerInfo 更新 Swagger 信息
func UpdateSwaggerInfo(config SwaggerConfig) {
	SwaggerInfo.Title = config.Title
	SwaggerInfo.Description = config.Description
	SwaggerInfo.Version = config.Version
	SwaggerInfo.Host = config.Host
	SwaggerInfo.BasePath = config.BasePath
	SwaggerInfo.Schemes = config.Schemes
}

// docTemplate Swagger 文档模板
const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "太上老君开发团队",
            "url": "https://github.com/taishanglaojun",
            "email": "dev@taishanglaojun.com"
        },
        "license": {
            "name": "MIT",
            "url": "https://opensource.org/licenses/MIT"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/auth/login": {
            "post": {
                "description": "用户登录",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "认证"
                ],
                "summary": "用户登录",
                "parameters": [
                    {
                        "description": "登录信息",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/LoginRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "登录成功",
                        "schema": {
                            "$ref": "#/definitions/LoginResponse"
                        }
                    },
                    "400": {
                        "description": "请求参数错误",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    },
                    "401": {
                        "description": "认证失败",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    }
                }
            }
        },
        "/auth/register": {
            "post": {
                "description": "用户注册",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "认证"
                ],
                "summary": "用户注册",
                "parameters": [
                    {
                        "description": "注册信息",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/RegisterRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "注册成功",
                        "schema": {
                            "$ref": "#/definitions/RegisterResponse"
                        }
                    },
                    "400": {
                        "description": "请求参数错误",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    },
                    "409": {
                        "description": "用户已存在",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    }
                }
            }
        },
        "/auth/refresh": {
            "post": {
                "description": "刷新访问令牌",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "认证"
                ],
                "summary": "刷新令牌",
                "parameters": [
                    {
                        "description": "刷新令牌",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/RefreshTokenRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "刷新成功",
                        "schema": {
                            "$ref": "#/definitions/RefreshTokenResponse"
                        }
                    },
                    "401": {
                        "description": "令牌无效",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    }
                }
            }
        },
        "/users": {
            "get": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "获取用户列表",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "用户管理"
                ],
                "summary": "获取用户列表",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "页码",
                        "name": "page",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "description": "每页数量",
                        "name": "limit",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "搜索关键词",
                        "name": "search",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "用户列表",
                        "schema": {
                            "$ref": "#/definitions/UserListResponse"
                        }
                    },
                    "401": {
                        "description": "未授权",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    }
                }
            },
            "post": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "创建新用户",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "用户管理"
                ],
                "summary": "创建用户",
                "parameters": [
                    {
                        "description": "用户信息",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/CreateUserRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "用户创建成功",
                        "schema": {
                            "$ref": "#/definitions/UserResponse"
                        }
                    },
                    "400": {
                        "description": "请求参数错误",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    },
                    "401": {
                        "description": "未授权",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    }
                }
            }
        },
        "/users/{id}": {
            "get": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "根据ID获取用户信息",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "用户管理"
                ],
                "summary": "获取用户信息",
                "parameters": [
                    {
                        "type": "string",
                        "description": "用户ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "用户信息",
                        "schema": {
                            "$ref": "#/definitions/UserResponse"
                        }
                    },
                    "401": {
                        "description": "未授权",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    },
                    "404": {
                        "description": "用户不存在",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    }
                }
            },
            "put": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "更新用户信息",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "用户管理"
                ],
                "summary": "更新用户",
                "parameters": [
                    {
                        "type": "string",
                        "description": "用户ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "用户信息",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/UpdateUserRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "用户更新成功",
                        "schema": {
                            "$ref": "#/definitions/UserResponse"
                        }
                    },
                    "400": {
                        "description": "请求参数错误",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    },
                    "401": {
                        "description": "未授权",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    },
                    "404": {
                        "description": "用户不存在",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    }
                }
            },
            "delete": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "删除用户",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "用户管理"
                ],
                "summary": "删除用户",
                "parameters": [
                    {
                        "type": "string",
                        "description": "用户ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "用户删除成功",
                        "schema": {
                            "$ref": "#/definitions/SuccessResponse"
                        }
                    },
                    "401": {
                        "description": "未授权",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    },
                    "404": {
                        "description": "用户不存在",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "LoginRequest": {
            "type": "object",
            "required": [
                "username",
                "password"
            ],
            "properties": {
                "username": {
                    "type": "string",
                    "example": "admin"
                },
                "password": {
                    "type": "string",
                    "example": "password123"
                }
            }
        },
        "LoginResponse": {
            "type": "object",
            "properties": {
                "access_token": {
                    "type": "string"
                },
                "refresh_token": {
                    "type": "string"
                },
                "expires_in": {
                    "type": "integer"
                },
                "user": {
                    "$ref": "#/definitions/User"
                }
            }
        },
        "RegisterRequest": {
            "type": "object",
            "required": [
                "username",
                "email",
                "password"
            ],
            "properties": {
                "username": {
                    "type": "string",
                    "example": "newuser"
                },
                "email": {
                    "type": "string",
                    "example": "user@example.com"
                },
                "password": {
                    "type": "string",
                    "example": "password123"
                },
                "full_name": {
                    "type": "string",
                    "example": "New User"
                }
            }
        },
        "RegisterResponse": {
            "type": "object",
            "properties": {
                "user": {
                    "$ref": "#/definitions/User"
                },
                "message": {
                    "type": "string"
                }
            }
        },
        "RefreshTokenRequest": {
            "type": "object",
            "required": [
                "refresh_token"
            ],
            "properties": {
                "refresh_token": {
                    "type": "string"
                }
            }
        },
        "RefreshTokenResponse": {
            "type": "object",
            "properties": {
                "access_token": {
                    "type": "string"
                },
                "expires_in": {
                    "type": "integer"
                }
            }
        },
        "User": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "string"
                },
                "username": {
                    "type": "string"
                },
                "email": {
                    "type": "string"
                },
                "full_name": {
                    "type": "string"
                },
                "avatar": {
                    "type": "string"
                },
                "role": {
                    "type": "string"
                },
                "is_active": {
                    "type": "boolean"
                },
                "created_at": {
                    "type": "string"
                },
                "updated_at": {
                    "type": "string"
                }
            }
        },
        "CreateUserRequest": {
            "type": "object",
            "required": [
                "username",
                "email",
                "password"
            ],
            "properties": {
                "username": {
                    "type": "string"
                },
                "email": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                },
                "full_name": {
                    "type": "string"
                },
                "role": {
                    "type": "string"
                }
            }
        },
        "UpdateUserRequest": {
            "type": "object",
            "properties": {
                "email": {
                    "type": "string"
                },
                "full_name": {
                    "type": "string"
                },
                "avatar": {
                    "type": "string"
                },
                "role": {
                    "type": "string"
                },
                "is_active": {
                    "type": "boolean"
                }
            }
        },
        "UserResponse": {
            "type": "object",
            "properties": {
                "user": {
                    "$ref": "#/definitions/User"
                }
            }
        },
        "UserListResponse": {
            "type": "object",
            "properties": {
                "users": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/User"
                    }
                },
                "pagination": {
                    "$ref": "#/definitions/Pagination"
                }
            }
        },
        "Pagination": {
            "type": "object",
            "properties": {
                "page": {
                    "type": "integer"
                },
                "limit": {
                    "type": "integer"
                },
                "total": {
                    "type": "integer"
                },
                "total_pages": {
                    "type": "integer"
                }
            }
        },
        "ErrorResponse": {
            "type": "object",
            "properties": {
                "error": {
                    "type": "string"
                },
                "message": {
                    "type": "string"
                },
                "code": {
                    "type": "integer"
                }
            }
        },
        "SuccessResponse": {
            "type": "object",
            "properties": {
                "message": {
                    "type": "string"
                },
                "success": {
                    "type": "boolean"
                }
            }
        }
    },
    "securityDefinitions": {
        "BearerAuth": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header",
            "description": "Bearer token authentication. Format: Bearer {token}"
        }
    }
}`
