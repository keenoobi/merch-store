import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
    scenarios: {
        constant_load: {
            executor: 'constant-arrival-rate', // Постоянная скорость запросов
            rate: 1000,                       // 1000 запросов в секунду
            timeUnit: '1s',                   // Единица времени — секунда
            duration: '2m',                   // Длительность теста — 5 минут
            preAllocatedVUs: 1000,            // Предварительно выделенные виртуальные пользователи
            maxVUs: 100000,                   // Максимальное количество виртуальных пользователей (100k)
        },
    },
    thresholds: {
        http_req_duration: ['p(95)<=50'], // 99.99% запросов должны выполняться за 50 мс или быстрее
        http_req_failed: ['rate<0.0001'],    // Частота ошибок должна быть меньше 0.01% (99.99% успешных запросов)
    },
};

export default function () {
    // Аутентификация
    let authRes = http.post('http://localhost:8080/api/auth', JSON.stringify({
        username: `user${__VU}`, // Уникальный username для каждого виртуального пользователя
        password: 'testpass',
    }), {
        headers: { 'Content-Type': 'application/json' },
    });

    // Проверка успешности аутентификации
    check(authRes, {
        'auth status is 200': (r) => r.status === 200,
    });

    // let token = authRes.json().token;

    // // // Получение информации о пользователе
    // let infoRes = http.get('http://localhost:8080/api/info', {
    //     headers: { 'Authorization': `Bearer ${token}` },
    // });

    // // Проверка успешности запроса информации
    // check(infoRes, {
    //     'info status is 200': (r) => r.status === 200,
    // });

    sleep(1); // Пауза между итерациями
}