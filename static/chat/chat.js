document.addEventListener('DOMContentLoaded', async () => {
    function getCookie(name) {
        const value = `; ${document.cookie}`;
        const parts = value.split(`; ${name}=`);
        if (parts.length === 2) return parts.pop().split(';').shift();
    }
    const username = getCookie('username');

    // if username cookie is not present, user does not have an access token
    if (username === undefined) {
        try {
            const response = await fetch('http://localhost:8080/api/refresh', {
                method: 'POST',
            });

            console.log(response.status)

            if (response.status === 401) {
                alert('Unauthorized, please log in to your account')
                window.location.href = 'http://localhost:8080/login';
            }
            if (response.status === 500) {
                alert('Internal server error')
                window.location.href = 'http://localhost:8080/login';
            }
        } catch (error) {
            alert('An error occurred. please log in to your account');
            window.location.href = 'http://localhost:8080/login';
        }
    }

    const socket = new WebSocket('ws://localhost:8080/api/chat');
    const chatWindow = document.getElementById('chat-window');
    const messageInput = document.getElementById('message-input');
    const sendButton = document.getElementById('send-button');

    socket.onerror = () => {
        alert('Not Authorized. Please login.');
        window.location.href = 'http://localhost:8080/login';
    };

    socket.onmessage = (event) => {
        const data = JSON.parse(event.data);
        console.log(data)
        addMessage(data.author, data.text);
    };

    sendButton.addEventListener('click', () => {
        sendMessage();
    });

    messageInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
            sendMessage();
        }
    });

    function sendMessage() {
        const text = messageInput.value.trim();
        if (text !== '') {
            socket.send(text);
            messageInput.value = '';
        }
    }

    function addMessage(author, text) {
        const messageElement = document.createElement('div');
        messageElement.classList.add('chat-message');
        messageElement.classList.add(author === username ? 'right' : 'left');
        messageElement.innerHTML = `<div class="sender-id">${author}</div><div class="text">${text}</div>`;
        chatWindow.appendChild(messageElement);
        chatWindow.scrollTop = chatWindow.scrollHeight;
    }
});
