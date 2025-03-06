package main

import (
	"bytes"
	"context"
	"fmt"
	"golang.org/x/image/draw"
	"image"
	"image/jpeg"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
)

const (
	qualidade = 75
	tamanho   = 1920
)

func main() {
	ctx := context.Background()

	endpoint := "localhost:9011"
	accessKeyID := "minio"        // Teu usuário do MinIO.
	secretAccessKey := "minio123" // Tua senha do MinIO.
	useSSL := false
	bucketName := "bucket-publico" // Nome do bucket.

	filePath := "./imagens/minha-imagem.jpg"

	objectName := "imagens/" + filepath.Base(filePath)

	// Conecta no MinIO.
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalf("Deu ruim na conexão com o MinIO: %v", err)
	}

	// Vê se o bucket já existe.
	exists, err := minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		log.Fatalf("Erro ao checar o bucket: %v", err)
	}

	// Se o bucket NÃO existir, cria ele.
	if !exists {
		err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			log.Fatalf("Erro ao criar o bucket: %v", err)
		}
		fmt.Printf("Bucket '%s' criado!\n", bucketName)
	}

	// Aqui a gente TORNAR O BUCKET PÚBLICO.  Cuidado! Isso libera geral pra ver o conteúdo.
	policy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::` + bucketName + `/*"]}]}`
	err = minioClient.SetBucketPolicy(ctx, bucketName, policy)
	if err != nil {
		log.Fatalf("Erro ao liberar o bucket: %v", err)
	}
	fmt.Printf("Bucket '%s' liberado pra geral!\n", bucketName)

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		log.Fatalf("Erro ao pegar info do arquivo: %v", err)
	}

	fileSize := fileInfo.Size()
	if fileSize == 0 {
		log.Fatalf("Vacilo! O arquivo '%s' tá vazio.", filePath)
	}

	fileSizeMB := float64(fileSize) / (1024 * 1024)
	fmt.Printf("Tamanho do arquivo: %.2f MB\n", fileSizeMB)

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Erro ao abrir a imagem: %v", err)
	}
	defer file.Close()

	img, err := jpeg.Decode(file)
	if err != nil {
		log.Fatalf("Erro ao decodificar a imagem: %v", err)
	}

	img = resizeImage(img, tamanho)

	// Converte para WebP usando a função
	webpBuffer, err := convertImageToWebPBuffer(img)
	if err != nil {
		log.Fatalf("Erro ao converter para WebP: %v", err)
	}

	contentType, err := getContentType(*webpBuffer)
	if err != nil {
		panic(err)
	}

	objectName = getWebPImagePath(objectName)

	_, err = minioClient.PutObject(ctx, bucketName, objectName, webpBuffer, int64(webpBuffer.Len()), minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		log.Fatalf("Erro no upload: %v", err)
	}
	fmt.Printf("Arquivo '%s' no bucket '%s'! Sucesso!\n", objectName, bucketName)

	// André, como o bucket é público, a gente MONTA a URL na mão. É sempre esse formato.
	objectURL := fmt.Sprintf("http://%s/%s/%s", endpoint, bucketName, objectName)
	fmt.Println("URL da imagem:", objectURL)

	resp, err := http.Get(objectURL)
	if err != nil {
		fmt.Println("Ih, a URL deu ruim:", err)
	} else {
		fmt.Println("URL respondeu:", resp.Status)
		resp.Body.Close()
	}

	// André: Gera um HTMLzinho pra mostrar a imagem. Copia e cola num arquivo .html pra ver.
	fmt.Printf("<img src=\"%s\" alt=\"Minha Imagem\">\n", objectURL)
}

// Converte image.Image para webp e retorna o buffer
func convertImageToWebPBuffer(img image.Image) (*bytes.Buffer, error) {
	var imagem bytes.Buffer

	options, err := encoder.NewLossyEncoderOptions(encoder.PresetDefault, qualidade)
	if err != nil {
		return nil, err
	}

	if err := webp.Encode(&imagem, img, options); err != nil {
		return nil, err
	}

	return &imagem, nil
}

func getWebPImagePath(imagePath string) string {
	outputName := strings.TrimSuffix(imagePath, filepath.Ext(imagePath)) + ".webp"
	return outputName
}

func getContentType(imagem bytes.Buffer) (string, error) {

	contentType := http.DetectContentType(imagem.Bytes())
	contentType = strings.Split(contentType, ";")[0]

	return contentType, nil
}

func resizeImage(img image.Image, maxWidth int) image.Image {
	originalWidth := img.Bounds().Dx()
	originalHeight := img.Bounds().Dy()

	if originalWidth > maxWidth {
		newHeight := (originalHeight * maxWidth) / originalWidth

		resizedImg := image.NewRGBA(image.Rect(0, 0, maxWidth, newHeight))

		draw.CatmullRom.Scale(resizedImg, resizedImg.Bounds(), img, img.Bounds(), draw.Over, nil)

		return resizedImg
	}

	return img
}
