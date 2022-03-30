
CREATE TABLE IF NOT EXISTS url(
    id SERIAL PRIMARY KEY not null,
    longurl VARCHAR(255) not null,
    shorturl VARCHAR(255),
    status VARCHAR(255)
);
---- create above / drop below ----
DROP TABLE url;