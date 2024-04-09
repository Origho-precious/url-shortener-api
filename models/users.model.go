package models

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"time"

	"github.com/Origho-precious/url-shortener/go/configs"
	"github.com/Origho-precious/url-shortener/go/services"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	Email         string
	FullName      string
	Password      string
	AuthToken     string
	CreatedAt     time.Time
	EmailVerified bool
}

type UserService struct {
	User                        User
	UserCollection              *mongo.Collection
	ForgotPasswordCollection    *mongo.Collection
	VerificationTokenCollection *mongo.Collection
}

func (us *UserService) hashPassword() ([]byte, error) {
	hashByte, err := bcrypt.GenerateFromPassword(
		[]byte(us.User.Password), bcrypt.DefaultCost,
	)

	return hashByte, err
}

func (us *UserService) comparePassword(hashedPassword []byte) error {
	err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(us.User.Password))
	return err
}

func (us *UserService) generateAuthToken() (string, error) {
	cfg, err := configs.LoadEnvs()
	if err != nil {
		log.Println("(1.) ", err)
		return "", err
	}

	jwtSecret := []byte(cfg.JWT_SECRET)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp":       time.Now().AddDate(0, 1, 0).Unix(),
		"userId":    us.User.ID.Hex(),
		"userEmail": us.User.Email,
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		log.Println("(2.) ", err)
		return "", err
	}

	return tokenString, nil
}

func (us *UserService) generateRandomToken() (string, error) {
	randomNumber, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "", err
	}

	number := int(randomNumber.Int64()) + 100000

	return strconv.Itoa(number), nil
}

func (us *UserService) sendEmailVerificationEmail() error {
	cfg, err := configs.LoadEnvs()
	if err != nil {
		log.Println(err)
		return fmt.Errorf("internal server error")
	}

	randNum, err := us.generateRandomToken()
	if err != nil {
		log.Println(err)
		return fmt.Errorf("internal server error")
	}

	// Save token to verificationToken collection
	_, err = us.VerificationTokenCollection.InsertOne(
		context.Background(), bson.M{
			"token":     randNum,
			"userId":    us.User.ID,
			"createdAt": time.Now(),
		},
	)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("internal server error")
	}

	emailService := services.Email{
		Title: "Email Verification",
		Extra: map[string]string{
			"user_email":        us.User.Email,
			"verification_code": randNum,
		},
		Recipients:   []string{us.User.Email},
		TemplateUUID: cfg.EMAIL_VERIFICATION_TEMPLATE_UUID,
	}

	msg, err := emailService.SendEmail()
	if err != nil {
		log.Println(err)
		return fmt.Errorf("internal server error")
	}

	log.Println(msg)

	return nil
}

func (us *UserService) CreateUser() (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Check if user with email already exists
	var existingUser User

	filter := bson.M{"email": us.User.Email}
	findErr := us.UserCollection.FindOne(ctx, filter).Decode(&existingUser)

	if findErr == mongo.ErrNoDocuments {
		hashByte, err := us.hashPassword()
		if err != nil {
			return User{}, err
		}

		passwordHash := string(hashByte)

		// Create User record in DB
		userRecord, err := us.UserCollection.InsertOne(ctx, bson.M{
			"email":         us.User.Email,
			"fullName":      us.User.FullName,
			"password":      passwordHash,
			"createdAt":     time.Now(),
			"emailVerified": false,
		})
		if err != nil {
			return User{}, err
		}

		insertedID, ok := userRecord.InsertedID.(primitive.ObjectID)
		if !ok {
			log.Println("Unexpected type for insertedID")
			return User{}, fmt.Errorf("internal server error")
		}

		us.User = User{
			ID:            insertedID,
			Email:         us.User.Email,
			FullName:      us.User.FullName,
			EmailVerified: false,
		}

		authToken, err := us.generateAuthToken()
		if err != nil {
			return User{}, fmt.Errorf("internal server error")
		}

		us.User.AuthToken = authToken

		go func() {
			err = us.sendEmailVerificationEmail()
			if err != nil {
				log.Println(err)
			}
		}()

		return us.User, nil
	} else if findErr != nil {
		return User{}, findErr
	}

	return User{}, fmt.Errorf(
		"user with email address: %s already exists", existingUser.Email,
	)
}

func (us *UserService) AuthenticateUser() (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var userData User

	filter := bson.M{"email": us.User.Email}

	err := us.UserCollection.FindOne(ctx, filter).Decode(&userData)
	if err == mongo.ErrNoDocuments {
		return User{}, fmt.Errorf("invalid email or password")
	} else if err != nil {
		return User{}, err
	}

	err = us.comparePassword([]byte(userData.Password))
	if err != nil {
		return User{}, fmt.Errorf("invalid email or password")
	}

	us.User.ID = userData.ID

	authToken, err := us.generateAuthToken()
	if err != nil {
		return User{}, fmt.Errorf("internal server error")
	}

	return User{
		ID:            userData.ID,
		Email:         userData.Email,
		FullName:      userData.FullName,
		AuthToken:     authToken,
		EmailVerified: userData.EmailVerified,
	}, nil
}

func (us *UserService) VerifyUserEmail(token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Find the token record
	var tokenRecord struct {
		Token  string
		ID     primitive.ObjectID `bson:"_id,omitempty"`
		UserID primitive.ObjectID `bson:"userId,omitempty"`
	}

	tokenFilter := bson.M{"userId": us.User.ID, "token": token}
	err := us.VerificationTokenCollection.FindOne(ctx, tokenFilter).Decode(
		&tokenRecord,
	)
	if err == mongo.ErrNoDocuments {
		log.Println(err)
		return fmt.Errorf("invalid token")
	} else if err != nil {
		log.Println(err)
		return err
	}

	// Update the user's emailVerified field to true
	var updatedUser User

	userFilter := bson.M{"_id": us.User.ID}
	update := bson.M{"$set": bson.M{"emailVerified": true}}
	options := options.FindOneAndUpdate().SetReturnDocument(options.After)

	err = us.UserCollection.FindOneAndUpdate(ctx, userFilter, update, options).Decode(&updatedUser)

	if err == mongo.ErrNoDocuments {
		return fmt.Errorf("internal server error")
	} else if err != nil {
		return err
	}

	// Delete the VerficationToken record
	tokenFilter = bson.M{"userId": us.User.ID, "token": token}
	err = us.VerificationTokenCollection.FindOneAndDelete(
		ctx, tokenFilter,
	).Decode(&tokenRecord)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (us *UserService) ResendEmailVerificationToken() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var verifTokenRecord struct {
		ID        primitive.ObjectID `bson:"_id,omitempty"`
		Token     string
		UserId    string
		createdAt time.Time
	}

	filter := bson.M{"userId": us.User.ID}
	err := us.VerificationTokenCollection.FindOne(ctx, filter).Decode(
		&verifTokenRecord,
	)
	if err == mongo.ErrNoDocuments {
		// Create new token and send
		err = us.sendEmailVerificationEmail()
		if err != nil {
			log.Println(err)
			return err
		}
	} else if err != nil {
		log.Println(err)
		return err
	}

	// Resend existing token
	cfg, err := configs.LoadEnvs()
	if err != nil {
		log.Println(err)
		return fmt.Errorf("internal server error")
	}

	emailService := services.Email{
		Title: "Email Verification",
		Extra: map[string]string{
			"user_email":        us.User.Email,
			"verification_code": verifTokenRecord.Token,
		},
		Recipients:   []string{us.User.Email},
		TemplateUUID: cfg.EMAIL_VERIFICATION_TEMPLATE_UUID,
	}

	msg, err := emailService.SendEmail()
	if err != nil {
		log.Println(err)
		return fmt.Errorf("internal server error")
	}

	log.Println(msg)

	return nil
}

func (us *UserService) ForgotPassword() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var existingUser User

	filter := bson.M{"email": us.User.Email}
	err := us.UserCollection.FindOne(ctx, filter).Decode(&existingUser)
	if err == mongo.ErrNoDocuments {
		return fmt.Errorf(
			"no account associated with email address: %s", us.User.Email,
		)
	} else if err != nil {
		return err
	}

	res, err := us.ForgotPasswordCollection.InsertOne(ctx, bson.M{
		"userEmail": us.User.Email,
		"createdAt": time.Now(),
	})
	if err != nil {
		return err
	}

	insertedID, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return fmt.Errorf("unexpected type for InsertedID")
	}

	cfg, err := configs.LoadEnvs()
	if err != nil {
		return err
	}

	resetPasswordURL := fmt.Sprintf(
		"%s?rspm=true&token=%s", cfg.CLIENT_URL, insertedID.Hex(),
	)

	emailService := services.Email{
		Title: "Password Reset",
		Extra: map[string]string{
			"pass_reset_link": resetPasswordURL,
			"user_email":      existingUser.Email,
		},
		Recipients:   []string{existingUser.Email},
		TemplateUUID: cfg.RESET_PASSWORD_TEMPLATE_UUID,
	}

	msg, err := emailService.SendEmail()
	if err != nil {
		return err
	}

	log.Println(msg)

	return nil
}

func (us *UserService) ResetPassword(token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var resetPasswordRecord struct {
		ID        primitive.ObjectID `bson:"_id,omitempty"`
		Token     string
		UserEmail string
		createdAt time.Time
	}

	objectID, err := primitive.ObjectIDFromHex(token)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("internal server error")
	}

	filter := bson.M{"_id": objectID}
	err = us.ForgotPasswordCollection.FindOne(ctx, filter).Decode(
		&resetPasswordRecord,
	)
	if err == mongo.ErrNoDocuments {
		log.Println(err)
		return fmt.Errorf("invalid token")
	} else if err != nil {
		log.Println(err)
		return err
	}

	hashedPassword, err := us.hashPassword()
	if err != nil {
		log.Println(err)
		return fmt.Errorf("internal server error")
	}

	filter = bson.M{"email": resetPasswordRecord.UserEmail}
	update := bson.M{"$set": bson.M{
		"password": string(hashedPassword), "emailVerified": true,
	}}
	res := us.UserCollection.FindOneAndUpdate(ctx, filter, update)
	if res.Err() != nil {
		log.Println(res.Err())
		return fmt.Errorf("internal server error")
	}

	go func() {
		filter = bson.M{"_id": objectID}
		res := us.ForgotPasswordCollection.FindOneAndDelete(
			context.Background(), filter,
		)
		if res.Err() != nil {
			log.Println(res.Err())
			return
		}

		log.Println("Forgot password record successfully deleted!")
	}()

	return nil
}

func (us *UserService) GetUser() (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var user User

	filter := bson.M{"_id": us.User.ID}
	err := us.UserCollection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		log.Println(err)
		return User{}, err
	}

	return User{
		ID:            user.ID,
		Email:         user.Email,
		FullName:      user.FullName,
		CreatedAt:     user.CreatedAt,
		EmailVerified: user.EmailVerified,
	}, nil
}

func (us *UserService) UpdateUserFullName() (User, error) {
	var user User

	filter := bson.M{"_id": us.User.ID}
	update := bson.M{"$set": bson.M{"fullName": us.User.FullName}}
	options := options.FindOneAndUpdate().SetReturnDocument(options.After)

	err := us.UserCollection.FindOneAndUpdate(
		context.TODO(), filter, update, options,
	).Decode(&user)

	if err != nil {
		log.Println(err)
		return User{}, err
	}

	return User{
		ID:            user.ID,
		Email:         user.Email,
		FullName:      user.FullName,
		CreatedAt:     user.CreatedAt,
		EmailVerified: user.EmailVerified,
	}, nil
}
