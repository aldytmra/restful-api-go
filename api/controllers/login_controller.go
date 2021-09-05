package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aldytmra/restful-api-go/api/auth"
	"github.com/aldytmra/restful-api-go/api/formaterror"
	"github.com/aldytmra/restful-api-go/api/models"
	"github.com/aldytmra/restful-api-go/api/responses"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

func (server *Server) Login(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	user := models.User{}
	err = json.Unmarshal(body, &user)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	user.Prepare()
	err = user.Validate("login")
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	token, err := server.SignIn(user.Email, user.Password)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusUnprocessableEntity, formattedError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    token["access_token"].(string),
		Expires:  time.Unix(token["exp_access_token"].(int64), 0),
		HttpOnly: true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    token["refresh_token"].(string),
		Expires:  time.Unix(token["exp_refresh_token"].(int64), 0),
		HttpOnly: true,
	})
	responses.JSON(w, http.StatusOK, token)
}

func (server *Server) Refresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := r.Cookie("refresh_token")
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	//verify the token
	token, err := jwt.Parse(refreshToken.Value, func(token *jwt.Token) (interface{}, error) {
		//Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("REFRESH_SECRET")), nil
	})
	//if there is an error, the token must have expired
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	//is token valid?
	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	//Since token is valid, get the uuid:
	claims, ok := token.Claims.(jwt.MapClaims) //the token claims should conform to MapClaims
	if ok && token.Valid {
		refreshUuid, ok := claims["refresh_uuid"].(string) //convert the interface to string
		if !ok {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
		userId, err := strconv.ParseUint(fmt.Sprintf("%.f", claims["user_id"]), 10, 64)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
		//Delete the previous Refresh Token
		deleted, delErr := auth.DeleteAuth(refreshUuid)
		if delErr != nil || deleted == 0 { //if any goes wrong
			responses.JSON(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		//Create new pairs of refresh and access tokens
		ts, createErr := auth.CreateToken(uint32(userId))
		if createErr != nil {
			responses.JSON(w, http.StatusForbidden, createErr.Error())
			return
		}
		//save the tokens metadata to redis
		saveErr := auth.CreateAuth(uint32(userId), ts)
		if saveErr != nil {
			responses.JSON(w, http.StatusForbidden, saveErr.Error())
			fmt.Println("saveErr", saveErr.Error())
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "access_token",
			Value:    ts.AccessToken,
			Expires:  time.Unix(ts.AtExpires, 0),
			HttpOnly: true,
		})
		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    ts.RefreshToken,
			Expires:  time.Unix(ts.RtExpires, 0),
			HttpOnly: true,
		})
		tokens := map[string]interface{}{
			"access_token":      ts.AccessToken,
			"exp_access_token":  ts.AtExpires,
			"refresh_token":     ts.RefreshToken,
			"exp_refresh_token": ts.RtExpires,
		}
		responses.JSON(w, http.StatusCreated, tokens)
	} else {
		responses.JSON(w, http.StatusUnauthorized, "refresh expired")
	}
}

func (server *Server) SignIn(email, password string) (map[string]interface{}, error) {

	var err error

	user := models.User{}

	err = server.DB.Debug().Model(models.User{}).Where("email = ?", email).Take(&user).Error
	mapEmpty := make(map[string]interface{})
	if err != nil {
		return mapEmpty, err
	}
	err = models.VerifyPassword(user.Password, password)
	fmt.Println("Verify Pass 1.0.10: ", err)
	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		return mapEmpty, err
	}

	ts, err := auth.CreateToken(user.ID)
	if err != nil {
		return mapEmpty, err
	}
	saveErr := auth.CreateAuth(user.ID, ts)
	if saveErr != nil {
		return mapEmpty, err
	}
	tokens := map[string]interface{}{
		"access_token":      ts.AccessToken,
		"exp_access_token":  ts.AtExpires,
		"refresh_token":     ts.RefreshToken,
		"exp_refresh_token": ts.RtExpires,
	}
	return tokens, nil
}
