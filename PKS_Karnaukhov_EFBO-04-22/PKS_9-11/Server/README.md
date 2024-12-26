# Скрипт для бд

```sql
-- Создаем таблицу пользователей
CREATE TABLE users (
    user_id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Создаем таблицу продуктов
CREATE TABLE products (
    product_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    stock INTEGER NOT NULL DEFAULT 0,
    image_url TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Создаем таблицу избранного
CREATE TABLE favorites (
    favorite_id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    product_id INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(product_id) ON DELETE CASCADE,
    UNIQUE(user_id, product_id)
);

-- Создаем таблицу корзины
CREATE TABLE cart (
    cart_id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    product_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(product_id) ON DELETE CASCADE,
    UNIQUE(user_id, product_id)
);

-- Создаем таблицу чатов
CREATE TABLE chats (
    chat_id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    admin_id INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (admin_id) REFERENCES administrators(admin_id) ON DELETE CASCADE
);

-- Создаем таблицу сообщений
CREATE TABLE messages (
    message_id SERIAL PRIMARY KEY,
    chat_id INTEGER NOT NULL,
    sender_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_read BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (chat_id) REFERENCES chats(chat_id) ON DELETE CASCADE,
    FOREIGN KEY (sender_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- Создаем индексы для оптимизации запросов
CREATE INDEX idx_favorites_user ON favorites(user_id);
CREATE INDEX idx_favorites_product ON favorites(product_id);
CREATE INDEX idx_cart_user ON cart(user_id);
CREATE INDEX idx_cart_product ON cart(product_id);
CREATE INDEX idx_messages_chat ON messages(chat_id);
CREATE INDEX idx_chats_user ON chats(user_id);
CREATE INDEX idx_chats_admin ON chats(admin_id);

-- Добавляем триггер для обновления updated_at в products
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_products_updated_at
    BEFORE UPDATE ON products
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Добавляем тестовые данные (опционально)
INSERT INTO users (username, email, password_hash) VALUES
('admin', 'admin@example.com', 'admin_hash'),
('user1', 'user1@example.com', 'user1_hash');

INSERT INTO administrators (user_id, role) VALUES
(1, 'super_admin');

INSERT INTO products (name, description, price, stock, image_url) VALUES
('Book 1', 'Description for Book 1', 29.99, 100, 'https://example.com/book1.jpg'),
('Book 2', 'Description for Book 2', 19.99, 50, 'https://example.com/book2.jpg');

-- Создаем таблицу заказов
CREATE TABLE orders (
    order_id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    total DECIMAL(10, 2) NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'new',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_status CHECK (status IN ('new', 'processing', 'shipped', 'delivered'))
);

-- Создаем таблицу для связи заказов с продуктами
CREATE TABLE order_products (
    order_id INTEGER REFERENCES orders(order_id) ON DELETE CASCADE,
    product_id INTEGER REFERENCES products(product_id) ON DELETE CASCADE,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    PRIMARY KEY (order_id, product_id)
);

-- Создаем индексы для оптимизации запросов
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_order_products_order_id ON order_products(order_id);

-- Создаем таблицу администраторов
CREATE TABLE IF NOT EXISTS administrators (
    admin_id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Создаем таблицу чатов
CREATE TABLE IF NOT EXISTS chats (
    chat_id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(user_id),
    admin_id INTEGER NOT NULL REFERENCES administrators(admin_id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Создаем таблицу сообщений
CREATE TABLE IF NOT EXISTS messages (
    message_id SERIAL PRIMARY KEY,
    chat_id INTEGER NOT NULL REFERENCES chats(chat_id),
    sender_type VARCHAR(10) NOT NULL CHECK (sender_type IN ('user', 'admin')),
    sender_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_read BOOLEAN DEFAULT FALSE
);

-- Обновляем таблицу чатов
DROP TABLE IF EXISTS chats CASCADE;
CREATE TABLE chats (
    chat_id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(user_id),
    admin_id INTEGER NOT NULL REFERENCES administrators(admin_id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Создаем новую таблицу сообщений
CREATE TABLE chat_messages (
    message_id SERIAL PRIMARY KEY,
    chat_id INTEGER NOT NULL REFERENCES chats(chat_id) ON DELETE CASCADE,
    sender_type VARCHAR(10) NOT NULL CHECK (sender_type IN ('user', 'admin')),
    sender_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_read BOOLEAN DEFAULT FALSE
);

-- Добавляем индексы
CREATE INDEX idx_chats_user_id ON chats(user_id);
CREATE INDEX idx_chats_admin_id ON chats(admin_id);
CREATE INDEX idx_chat_messages_chat_id ON chat_messages(chat_id);
CREATE INDEX idx_chat_messages_sender_id ON chat_messages(sender_id);

-- Создаем триггер для обновления updated_at в чатах
CREATE OR REPLACE FUNCTION update_chat_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE chats 
    SET updated_at = CURRENT_TIMESTAMP 
    WHERE chat_id = NEW.chat_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_chat_timestamp
    AFTER INSERT ON chat_messages
    FOR EACH ROW
    EXECUTE FUNCTION update_chat_timestamp();

-- Добавляем тестовые данные
DO $$
DECLARE
    v_user_id INTEGER;
    v_admin_id INTEGER;
    v_chat_id INTEGER;
BEGIN
    -- Получаем или создаем тестового пользователя
    SELECT user_id INTO v_user_id FROM users LIMIT 1;
    IF v_user_id IS NULL THEN
        INSERT INTO users (username, email, password_hash)
        VALUES ('test_user', 'test@example.com', 'test_hash')
        RETURNING user_id INTO v_user_id;
    END IF;

    -- Получаем или создаем администратора
    SELECT admin_id INTO v_admin_id FROM administrators LIMIT 1;
    IF v_admin_id IS NULL THEN
        INSERT INTO administrators (username, email, password_hash)
        VALUES ('admin', 'admin@example.com', '$2a$06$bCvGwBVcvsosTCsEtoG60.nspP/RIUgKgZGi9TYd/ZwWllr6UTeyK')
        RETURNING admin_id INTO v_admin_id;
    END IF;

    -- Создаем тестовый чат
    INSERT INTO chats (user_id, admin_id)
    VALUES (v_user_id, v_admin_id)
    RETURNING chat_id INTO v_chat_id;

    -- Добавляем тестовые сообщения
    INSERT INTO chat_messages (chat_id, sender_type, sender_id, content, is_read) VALUES
    (v_chat_id, 'admin', v_admin_id, 'Здравствуйте! Чем могу помочь?', true),
    (v_chat_id, 'user', v_user_id, 'У меня вопрос по заказу', true),
    (v_chat_id, 'admin', v_admin_id, 'Конечно, какой у вас вопрос?', false);
END $$;

-- Создаем полезные функции
CREATE OR REPLACE FUNCTION get_chat_messages(p_chat_id INTEGER)
RETURNS TABLE (
    message_id INTEGER,
    chat_id INTEGER,
    sender_type VARCHAR,
    sender_id INTEGER,
    content TEXT,
    sent_at TIMESTAMP,
    is_read BOOLEAN
) AS $$
BEGIN
    RETURN QUERY
    SELECT m.message_id, m.chat_id, m.sender_type, m.sender_id, 
           m.content, m.sent_at, m.is_read
    FROM chat_messages m
    WHERE m.chat_id = p_chat_id
    ORDER BY m.sent_at ASC;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION get_user_chats(p_user_id INTEGER)
RETURNS TABLE (
    chat_id INTEGER,
    admin_name VARCHAR,
    last_message TEXT,
    last_message_time TIMESTAMP,
    unread_count INTEGER
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        c.chat_id,
        a.username as admin_name,
        m.content as last_message,
        m.sent_at as last_message_time,
        COUNT(um.message_id)::INTEGER as unread_count
    FROM chats c
    JOIN administrators a ON a.admin_id = c.admin_id
    LEFT JOIN LATERAL (
        SELECT content, sent_at
        FROM chat_messages
        WHERE chat_id = c.chat_id
        ORDER BY sent_at DESC
        LIMIT 1
    ) m ON true
    LEFT JOIN chat_messages um ON 
        um.chat_id = c.chat_id AND 
        um.is_read = false AND 
        um.sender_type = 'admin'
    WHERE c.user_id = p_user_id
    GROUP BY c.chat_id, a.username, m.content, m.sent_at
    ORDER BY m.sent_at DESC NULLS LAST;
END;
$$ LANGUAGE plpgsql;
```