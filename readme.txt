curl для получения новостей с фильтрацией и пагинацией
curl -X GET "http://localhost:8080/news/filter?s=провела&page=1"

curl для получения новостей и для получения новостей с пагинацией
curl -X GET "http://localhost:8080/news"
curl -X GET "http://localhost:8080/news?page=3"

curl для получения полной информации о новости с комментариями
curl -X GET http://localhost:8080/news/1

curl для добавления комментария к новости
curl -X POST http://localhost:8080/news/1/comment -H "Content-Type: application/json" -d "{\"text\": \"Отличная статья!\", \"parent_id\": null}"