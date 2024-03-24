package models

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"github.com/Origho-precious/url-shortener/go/configs"
	"github.com/Origho-precious/url-shortener/go/services"
	"github.com/imagekit-developer/imagekit-go"
	"github.com/imagekit-developer/imagekit-go/api/uploader"
	"github.com/skip2/go-qrcode"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Url struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	UserId         primitive.ObjectID `bson:"userId,omitempty"`
	Deleted        bool
	CreatedAt      time.Time
	ExpiresAt      time.Time
	VisitCount     int64
	CustomAlias    bool
	OriginalUrl    string
	ShortUrlSlug   string
	LastVisitedAt  time.Time
	QRCodeImageUrl string
}

type Visit struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	UrlId      primitive.ObjectID
	Browser    string
	Location   string
	Referrer   string
	IPAddress  string
	VisitedAt  time.Time
	DeviceType string
}

type UrlService struct {
	Url             Url
	Visit           Visit
	UrlCollection   *mongo.Collection
	VisitCollection *mongo.Collection
}

func (urlS *UrlService) getShortUrlSlug(alias string) error {
	if alias == "" {
		urlSlug := services.GenerateShortURL(urlS.Url.OriginalUrl)
		urlS.Url.ShortUrlSlug = urlSlug
		urlS.Url.CustomAlias = false

		return nil
	}

	var urlRecord Url

	filter := bson.M{"shortUrlSlug": alias}
	err := urlS.UrlCollection.FindOne(context.TODO(), filter).Decode(&urlRecord)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			urlS.Url.ShortUrlSlug = alias
			urlS.Url.CustomAlias = true

			return nil
		}

		return fmt.Errorf("internal server error")
	}

	return fmt.Errorf("url with this alias already exist")
}

func (urlS *UrlService) createQRCode(shortUrl string) (string, error) {
	qrCode, err := qrcode.Encode(shortUrl, qrcode.Medium, 256)
	if err != nil {
		return "", err
	}

	qrCodeBase64 := base64.StdEncoding.EncodeToString(qrCode)

	return qrCodeBase64, nil
}

func (urlS *UrlService) generateAndUploadQRCode() (string, error) {
	cfg, err := configs.LoadEnvs()
	if err != nil {
		return "", err
	}

	shortUrl := fmt.Sprintf(
		"%s/%s", cfg.URL_REDIRECT_PREFIX, urlS.Url.ShortUrlSlug,
	)

	base64Image, err := urlS.createQRCode(shortUrl)
	if err != nil {
		return "", err
	}

	ik := imagekit.NewFromParams(imagekit.NewParams{
		PublicKey:   cfg.IMAGEKIT_PUBLIC_KEY,
		PrivateKey:  cfg.IMAGEKIT_PRIVATE_KEY,
		UrlEndpoint: cfg.IMAGEKIT_URL_ENDPOINT,
	})

	response, err := ik.Uploader.Upload(
		context.TODO(), base64Image, uploader.UploadParam{
			FileName: urlS.Url.ShortUrlSlug,
			Folder:   "url-shortener",
		},
	)
	if err != nil {
		return "", err
	}

	return response.Data.Url, nil
}

func (urlS *UrlService) CreateShortUrl(alias string) (
	map[string]string, error,
) {
	err := urlS.getShortUrlSlug(alias)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	qrCodeUrl, err := urlS.generateAndUploadQRCode()
	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("internal server error")
	}

	insertedRecord, err := urlS.UrlCollection.InsertOne(context.TODO(), bson.M{
		"userId":         urlS.Url.UserId,
		"deleted":        false,
		"createdAt":      time.Now(),
		"expiresAt":      urlS.Url.ExpiresAt,
		"visitCount":     0,
		"customAlias":    urlS.Url.CustomAlias,
		"originalUrl":    urlS.Url.OriginalUrl,
		"shortUrlSlug":   urlS.Url.ShortUrlSlug,
		"lastVisitedAt":  time.Time{},
		"qrCodeImageUrl": qrCodeUrl,
	})
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("internal server error")
	}

	cfg, err := configs.LoadEnvs()
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("internal server error")
	}

	shortUrl := fmt.Sprintf(
		"%s/%s", cfg.URL_REDIRECT_PREFIX, urlS.Url.ShortUrlSlug,
	)

	id, ok := insertedRecord.InsertedID.(primitive.ObjectID)
	if !ok {
		log.Println("Unexpected type for insertedID")
		return nil, fmt.Errorf("internal server error")
	}

	res := map[string]string{
		"id":             id.Hex(),
		"shortUrl":       shortUrl,
		"originalUrl":    urlS.Url.OriginalUrl,
		"qrCodeImageUrl": qrCodeUrl,
	}

	return res, nil
}

func (urlS *UrlService) GetOriginalUrl() (Url, error) {
	var urlRecord Url

	filter := bson.M{"shortUrlSlug": urlS.Url.ShortUrlSlug, "deleted": false}
	err := urlS.UrlCollection.FindOne(context.TODO(), filter).Decode(&urlRecord)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Println(err)
			return Url{}, fmt.Errorf("invalid shorturl")
		}

		fmt.Println(err)
		return Url{}, fmt.Errorf("internal server error")
	}

	if !urlRecord.ExpiresAt.IsZero() && urlRecord.ExpiresAt.Before(time.Now()) {
		return Url{}, fmt.Errorf("url has expired")
	}

	return urlRecord, nil
}

func (urlS *UrlService) DeleteUrl() error {
	filter := bson.M{
		"_id":     urlS.Url.ID,
		"userId":  urlS.Url.UserId,
		"deleted": false,
	}
	update := bson.M{"$set": bson.M{"deleted": true}}
	result := urlS.UrlCollection.FindOneAndUpdate(context.TODO(), filter, update)
	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			return fmt.Errorf("no matching document found")
		}

		fmt.Println(result.Err())
		return fmt.Errorf("internal server error")
	}

	return nil
}

func (urlS *UrlService) GetUrlsByUser(page int, limit int) (
	[]Url, int64, error,
) {
	skip := (page - 1) * limit

	sort := bson.M{"createdAt": -1}
	opts := options.Find().SetSort(sort).SetSkip(
		int64(skip),
	).SetLimit(int64(limit))

	filter := bson.M{"userId": urlS.Url.UserId, "deleted": false}

	cursor, err := urlS.UrlCollection.Find(context.TODO(), filter, opts)
	if err != nil {
		fmt.Println(err)
		return []Url{}, 0, fmt.Errorf("internal server error")
	}

	var urlRecords []Url
	if err = cursor.All(context.TODO(), &urlRecords); err != nil {
		log.Println(err)
		return []Url{}, 0, fmt.Errorf("internal server error")
	}
	defer cursor.Close(context.TODO())

	var urls []Url
	for _, urlRecord := range urlRecords {
		urls = append(urls, Url{
			ID:             urlRecord.ID,
			CreatedAt:      urlRecord.CreatedAt,
			ExpiresAt:      urlRecord.ExpiresAt,
			VisitCount:     urlRecord.VisitCount,
			OriginalUrl:    urlRecord.OriginalUrl,
			CustomAlias:    urlRecord.CustomAlias,
			ShortUrlSlug:   urlRecord.ShortUrlSlug,
			LastVisitedAt:  urlRecord.LastVisitedAt,
			QRCodeImageUrl: urlRecord.QRCodeImageUrl,
		})
	}

	total, err := urlS.UrlCollection.CountDocuments(context.TODO(), filter)
	if err != nil {
		log.Println(err)
		return nil, 0, fmt.Errorf("internal server error")
	}

	return urls, total, nil
}

func (urlS *UrlService) SaveClickAnalytics() error {
	_, err := urlS.VisitCollection.InsertOne(context.TODO(), bson.M{
		"urlId":      urlS.Visit.UrlId,
		"browser":    urlS.Visit.Browser,
		"location":   urlS.Visit.Location,
		"referrer":   urlS.Visit.Referrer,
		"ipAddress":  urlS.Visit.IPAddress,
		"visitedAt":  urlS.Visit.VisitedAt,
		"deviceType": urlS.Visit.DeviceType,
	})

	if err != nil {
		return err
	}

	filter := bson.M{"_id": urlS.Visit.UrlId}
	update := bson.M{
		"$inc": bson.M{"visitCount": 1},
		"$set": bson.M{"lastVisitedAt": time.Now()},
	}
	res := urlS.UrlCollection.FindOneAndUpdate(context.TODO(), filter, update)
	if res.Err() != nil {
		return err
	}

	return nil
}
