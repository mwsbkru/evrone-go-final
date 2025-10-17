docker compose build && docker compose up

{
"user_email": "w1@rty.ru",
"subject": "Hello",
"body": "я пришел к тебе с приветом"
}

http://localhost:8080/notifications/subscribe?userEmail=w1@rty.ru

кейсы
- отлет клиента
- закрытие коннекта клиентом
- закрытие коннеккта сервером
- отправка сообщения когда пользователь оффлайн
- повторное подключение клиента с одним и тем же email
