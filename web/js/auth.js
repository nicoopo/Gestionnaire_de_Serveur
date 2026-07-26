const token = localStorage.getItem('token');
if (!token) {
    window.location.href = '/login.html';
}

function authFetch(url, options = {}) {
    options.headers = {
        ...(options.headers || {}),
        'Authorization': `Bearer ${token}`
    };
    return fetch(url, options).then(res => {
        if (res.status === 401) {
            localStorage.removeItem('token');
            window.location.href = '/login.html';
        }
        return res;
    });
}
