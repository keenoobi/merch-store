import http from 'k6/http';
import { check } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';
import { textSummary } from 'https://jslib.k6.io/k6-summary/0.0.1/index.js';

// Кастомные метрики для авторизации
const authRequests = new Counter('auth_requests'); // Количество запросов к /api/auth
const authSuccessRate = new Rate('auth_success_rate'); // Процент успешных запросов к /api/auth
const authResponseTime = new Trend('auth_response_time'); // Время ответа для /api/auth

// Конфигурация теста
export const options = {
    stages: [
        { duration: '15s', target: 250 },  // Постепенно увеличиваем нагрузку до 50 пользователей за 15 секунд
        { duration: '30s', target: 500 }, // Увеличиваем до 100 пользователей за 30 секунд
        { duration: '15s', target: 0 },   // Постепенно снижаем нагрузку до 0
    ],
    thresholds: {
        'auth_success_rate': ['rate>0.99'], // 95% запросов к /api/auth должны быть успешными
        'auth_response_time': ['p(99)<50'], // 95% запросов к /api/auth должны завершаться за 500 мс
    },
};

// Функция для логирования ошибок
function logError(endpoint, response) {
    if (response.status !== 200) {
        console.error(`Error in ${endpoint}: Status=${response.status}, Body=${response.body}`);
    }
}

// Основной сценарий тестирования
export default function () {
    const url = 'http://localhost:8080/api/auth';
    const payload = JSON.stringify({
        username: `user${__VU}`, // Уникальное имя пользователя для каждого виртуального пользователя
        password: 'testpassword',
    });
    const params = {
        headers: {
            'Content-Type': 'application/json',
        },
    };

    // Выполняем запрос на авторизацию
    const res = http.post(url, payload, params);

    // Проверяем статус ответа
    check(res, {
        'status is 200': (r) => r.status === 200,
    });

    // Обновляем метрики
    authRequests.add(1); // Увеличиваем счетчик запросов
    authSuccessRate.add(res.status === 200); // Учитываем успешный запрос
    authResponseTime.add(res.timings.duration); // Записываем время ответа

    // Логируем ошибки
    logError('/api/auth', res);
}

// Функция для сохранения отчета в файл
export function handleSummary(data) {
    return {
        'summary.json': JSON.stringify(data), // Сохраняем отчет в формате JSON
        'summary.txt': textSummary(data, { indent: ' ', enableColors: false }), // Сохраняем отчет в текстовом формате
    };
}