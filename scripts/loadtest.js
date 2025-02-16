import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';
import { textSummary } from 'https://jslib.k6.io/k6-summary/0.0.1/index.js';

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
        { duration: '15s', target: 50 },  // Постепенно увеличиваем нагрузку до 50 пользователей за 15 секунд
        { duration: '30s', target: 100 }, // Увеличиваем до 100 пользователей за 30 секунд
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

// Функция для логирования ошибок
function logError(endpoint, response) {
    if (response.status !== 200) {
        console.error(`Error in ${endpoint}: Status=${response.status}, Body=${response.body}`);
    }
}

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
    logError('/api/auth', res); // Логируем ошибки
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
    logError('/api/info', infoRes); // Логируем ошибки

    // Тестируем эндпоинт /api/buy/{item}
    const buyUrl = 'http://localhost:8080/api/buy/t-shirt';
    const buyRes = http.get(buyUrl, infoParams);
    check(buyRes, {
        'status is 200': (r) => r.status === 200,
    });
    buyRequests.add(1); // Увеличиваем счетчик запросов
    buySuccessRate.add(buyRes.status === 200); // Учитываем успешный запрос
    buyResponseTime.add(buyRes.timings.duration); // Записываем время ответа
    logError('/api/buy/{item}', buyRes); // Логируем ошибки

    // Пауза между запросами
    sleep(1);
}

// Функция для сохранения отчета в файл
export function handleSummary(data) {
    return {
        'summary.json': JSON.stringify(data), // Сохраняем отчет в формате JSON
        'summary.txt': textSummary(data, { indent: ' ', enableColors: false }), // Сохраняем отчет в текстовом формате
    };
}