document.addEventListener('DOMContentLoaded', function() {
    const form = document.querySelector('form');

    form.addEventListener('submit', async function(event) {
        event.preventDefault();

        const username = form.querySelector('input[name="username"]').value;
        const password = form.querySelector('input[name="password"]').value;

        const usernameError = validateUsername(username);
        const passwordError = validatePassword(password);

        if (usernameError) {
            alert(`Username Error: ${usernameError}`);
            return;
        }

        if (passwordError) {
            alert(`Password Error: ${passwordError}`);
            return;
        }

        try {
            const response = await fetch('https://chat-hub.liara.run/api/signup', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ username, password })
            });

            if (!response.ok) {
                const errorData = await response.json();
                alert(`Error: ${errorData.error}`);
                return;
            }

            window.location.href = 'https://chat-hub.liara.run/chat';
        } catch (error) {
            console.error('Error:', error);
            alert('An error occurred during signup. Please try again.');
        }
    });

    function validateUsername(username) {
        if (username.length < 4) {
            return "Username must be at least 4 characters.";
        }
        if (username.length > 64) {
            return "Username must be at most 64 characters.";
        }
        if (!/^[a-zA-Z][a-zA-Z0-9_]*$/.test(username)) {
            return "Username must contain only alphabets, digits, and underscore, and must start with an alphabet.";
        }
        return null;
    }

    function validatePassword(password) {
        if (password.length < 8) {
            return "Password must be at least 8 characters.";
        }
        if (password.length > 64) {
            return "Password must be at most 64 characters.";
        }
        if (!/^[a-zA-Z0-9_!@#$%&*^.]*$/.test(password)) {
            return "Invalid character in password; only alphabets, digits, and the following special characters are allowed: _!@#$%&*.^";
        }
        return null;
    }
});