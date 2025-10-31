# REST API для управления задачами

REST API для управления задачами (Task Manager).

1. **Middleware**:
   - Логирование запросов
   - Recovery от паники
   - CORS
   - Аутентификация (проверка session/JWT)

2. **База данных**: SQLite, PostgreSQL или in-memory
3. **Валидация**: Проверка входных данных
4. **Ошибки**: Правильные HTTP статус-коды и JSON ответы
5. **Health check endpoint**: `GET /health`