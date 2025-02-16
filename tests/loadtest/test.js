import http from 'k6/http';
import { check, sleep } from 'k6';
import { randomIntBetween } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

export let options = {
    scenarios: {
        // Этап создания пользователей
        create_users: {
            executor: 'shared-iterations',
            vus: 1,
            iterations: 100000,
            maxDuration: '5m',
            exec: 'createUser',
        },
        // Этап покупок
        purchase: {
            executor: 'constant-arrival-rate',
            startTime: '2m',
            rate: 1000,
            timeUnit: '1s',
            duration: '1m',
            preAllocatedVUs: 1000,
            maxVUs: 100000,
            exec: 'purchaseItem',
        },
        // Этап передачи монет
        send_coins: {
            executor: 'constant-arrival-rate',
            startTime: '3m',
            rate: 1000,
            timeUnit: '1s',
            duration: '1m',
            preAllocatedVUs: 1000,
            maxVUs: 100000,
            exec: 'sendCoins',
        },
        get_info: {
            executor: 'constant-arrival-rate',
            startTime: '4m',
            rate: 1000,
            timeUnit: '1s',
            duration: '1m',
            preAllocatedVUs: 1000,
            maxVUs: 100000,
            exec: 'getInfo',
        },
    },
    thresholds: {
        http_req_duration: ['p(95)<=50'],  // 95% запросов ≤ 50 мс
        http_req_failed: ['rate<0.0001'],  // Ошибки < 0.01% (99.99% успешных)
    },
};

// Функция для создания пользователей
export function createUser() {
    let username = `user${__ITER}`; // Уникальный username

    let authRes = http.post('http://avito-shop-service:8080/api/auth', JSON.stringify({
        username: username,
        password: 'testpass',
    }), {
        headers: { 'Content-Type': 'application/json' },
    });

    check(authRes, { 'auth status is 200': (r) => r.status === 200 });
}

// Функция для покупок
export function purchaseItem() {
    // Авторизация
    let authRes = http.post('http://avito-shop-service:8080/api/auth', JSON.stringify({
        username: `user${randomIntBetween(0, 9999)}`,
        password: 'testpass',
    }), {
        headers: { 'Content-Type': 'application/json' },
    });

    check(authRes, { 'auth status is 200': (r) => r.status === 200 });

    let token = authRes.json().token;

    // Покупка предмета
    let items = ['t-shirt', 'cup', 'book', 'pen', 'powerbank', 'hoody', 'umbrella', 'socks', 'wallet', 'pink-hoody'];
    let item = items[randomIntBetween(0, items.length - 1)];

    let buyRes = http.get(`http://avito-shop-service:8080/api/buy/${item}`, {
        headers: { 'Authorization': `Bearer ${token}` },
    });

    check(buyRes, { 'buy status is 200': (r) => r.status === 200 });
}

// Функция для передачи монет
export function sendCoins() {
    // Авторизация
    let authRes = http.post('http://avito-shop-service:8080/api/auth', JSON.stringify({
        username: `user${randomIntBetween(0, 49999)}`,
        password: 'testpass',
    }), {
        headers: { 'Content-Type': 'application/json' },
    });

    check(authRes, { 'auth status is 200': (r) => r.status === 200 });

    let token = authRes.json().token;

    // Отправка монет
    let sendCoinRes = http.post('http://avito-shop-service:8080/api/sendCoin', JSON.stringify({
        toUser: `user${randomIntBetween(50000, 99999)}`,
        amount: randomIntBetween(1, 100),
    }), {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`,
        },
    });

    check(sendCoinRes, { 'sendCoin status is 200': (r) => r.status === 200 });
}

export function getInfo() {
    // Авторизация
    let authRes = http.post('http://avito-shop-service:8080/api/auth', JSON.stringify({
        username: `user${randomIntBetween(0, 99999)}`,
        password: 'testpass',
    }), {
        headers: { 'Content-Type': 'application/json' },
    });

    check(authRes, { 'auth status is 200': (r) => r.status === 200 });

    let token = authRes.json().token;

    // Информация о пользователе
    let userInfo = http.get(`http://avito-shop-service:8080/api/info`, {
        headers: { 'Authorization': `Bearer ${token}` },
    });

    check(userInfo, { 'userInfo status is 200': (r) => r.status === 200 });
}
