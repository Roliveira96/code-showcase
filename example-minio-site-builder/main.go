package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
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

	// --- UPLOAD da imagem ---
	contentType := "image/jpeg"

	// André, a função FPutObject() é que manda a imagem pro MinIO.
	_, err = minioClient.FPutObject(ctx, bucketName, objectName, filePath, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		log.Fatalf("Erro no upload: %v", err)
	}
	fmt.Printf("Arquivo '%s' no bucket '%s'! Sucesso!\n", objectName, bucketName)

	// --- Monta a URL manualmente ---
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
