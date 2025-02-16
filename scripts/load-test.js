// import http from 'k6/http';
// import { check, sleep } from 'k6';

// export let options = {
//     stages: [
//         { duration: '1m', target: 100 },  // 100 пользователей в течение 1 минуты
//         { duration: '2m', target: 500 },  // 500 пользователей в течение 2 минут
//         { duration: '3m', target: 1000 }, // 1000 пользователей в течение 3 минут
//     ],
//     thresholds: {
//         'http_req_duration': ['p(99)<50'],  // 95% запросов должны быть меньше 50ms
//         'http_req_failed': ['rate<0.0001'],  // Ошибок должно быть меньше 0.01%
//     },
// };

// export default function () {
//     let res = http.post('http://avito-shop-service:8080/api/auth', JSON.stringify({
//         username: 'testuser',
//         password: 'testpass',
//     }), {
//         headers: { 'Content-Type': 'application/json' },
//     });

//     check(res, {
//         'status is 200': (r) => r.status === 200,
//     });

//     sleep(1);  // Пауза в 1 секунду между запросами
// }

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';

// Кастомные метрики для каждого эндпоинта
const authRequests = new Counter('auth_requests'); // Количество запросов к /api/auth
const authSuccessRate = new Rate('auth_success_rate'); // Процент успешных запросов к /api/auth
const authResponseTime = new Trend('auth_response_time'); // Время ответа для /api/auth

const infoRequests = new Counter('info_requests'); // Количество запросов к /api/info
const infoSuccessRate = new Rate('info_success_rate'); // Процент успешных запросов к /api/info
const infoResponseTime = new Trend('info_response_time'); // Время ответа для /api/info

const buyRequests = new Counter('buy_requests'); // Количество запросов к /api/buy/{item}
const buySuccessRate = new Rate('buy_success_rate'); // Процент успешных запросов к /api/buy/{item}
const buyResponseTime = new Trend('buy_response_time'); // Время ответа для /api/buy/{item}

// Конфигурация теста
export const options = {
    stages: [
        { duration: '15s', target: 1000 },  // Постепенно увеличиваем нагрузку до 50 пользователей за 15 секунд
        { duration: '30s', target: 1000 }, // Увеличиваем до 100 пользователей за 30 секунд
        { duration: '15s', target: 0 },   // Постепенно снижаем нагрузку до 0
    ],
    thresholds: {
        'auth_success_rate': ['rate>0.95'], // 95% запросов к /api/auth должны быть успешными
        'info_success_rate': ['rate>0.95'], // 95% запросов к /api/info должны быть успешными
        'buy_success_rate': ['rate>0.95'],  // 95% запросов к /api/buy/{item} должны быть успешными
        'auth_response_time': ['p(95)<500'], // 95% запросов к /api/auth должны завершаться за 500 мс
        'info_response_time': ['p(95)<500'], // 95% запросов к /api/info должны завершаться за 500 мс
        'buy_response_time': ['p(95)<500'],  // 95% запросов к /api/buy/{item} должны завершаться за 500 мс
    },
};

// Функция для аутентификации и получения токена
function authenticate() {
    const url = 'http://localhost:8080/api/auth';
    const payload = JSON.stringify({
        username: `user${__VU}`,
        password: 'testpassword',
    });
    const params = {
        headers: {
            'Content-Type': 'application/json',
        },
    };
    const res = http.post(url, payload, params);
    check(res, {
        'status is 200': (r) => r.status === 200,
    });
    authRequests.add(1); // Увеличиваем счетчик запросов
    authSuccessRate.add(res.status === 200); // Учитываем успешный запрос
    authResponseTime.add(res.timings.duration); // Записываем время ответа

    return res.json().token;
}

// Основной сценарий тестирования
export default function () {
    const token = authenticate();

    // Тестируем эндпоинт /api/info
    const infoUrl = 'http://localhost:8080/api/info';
    const infoParams = {
        headers: {
            Authorization: `Bearer ${token}`,
        },
    };
    const infoRes = http.get(infoUrl, infoParams);
    check(infoRes, {
        'status is 200': (r) => r.status === 200,
    });
    infoRequests.add(1); // Увеличиваем счетчик запросов
    infoSuccessRate.add(infoRes.status === 200); // Учитываем успешный запрос
    infoResponseTime.add(infoRes.timings.duration); // Записываем время ответа


    // Тестируем эндпоинт /api/buy/{item}
    const buyUrl = 'http://localhost:8080/api/buy/t-shirt';
    const buyRes = http.get(buyUrl, infoParams);
    check(buyRes, {
        'status is 200': (r) => r.status === 200,
    });
    buyRequests.add(1); // Увеличиваем счетчик запросов
    buySuccessRate.add(buyRes.status === 200); // Учитываем успешный запрос
    buyResponseTime.add(buyRes.timings.duration); // Записываем время ответа


    // Пауза между запросами
    sleep(1);
}