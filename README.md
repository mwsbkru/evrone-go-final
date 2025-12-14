docker compose build && docker compose up
clear && GOEXPERIMENT=synctest go test ./...


{
"user_email": "w1@rty.ru",
"subject": "Hello",
"body": "я пришел к тебе с приветом"
}

http://localhost:8080/notifications/subscribe?userEmail=w1@rty.ru


- http://localhost:8000 - kafka
- http://localhost:8025/ - mailhog
- http://localhost:5540/ - redis

кейсы
- отлет клиента
- закрытие коннекта клиентом
- закрытие коннеккта сервером
- отправка сообщения когда пользователь оффлайн
- повторное подключение клиента с одним и тем же email
