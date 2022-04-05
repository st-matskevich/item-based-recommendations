package firebase

import (
	"context"
	"os"

	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

func OpenFirebaseClient() error {
	opt := option.WithCredentialsJSON([]byte(os.Getenv("FIREBASE_JSON")))
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return err
	}

	auth, err := app.Auth(context.Background())
	if err == nil {
		authClient = &FirebaseAuth{fbAuth: auth}
	}

	return err
}
