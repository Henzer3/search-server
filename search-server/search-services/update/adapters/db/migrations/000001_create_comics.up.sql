CREATE TABLE IF NOT EXISTS comics (
    num SERIAL PRIMARY KEY,
    img_url TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS words (
    id SERIAL PRIMARY KEY,
    word TEXT NOT NULL,
    comics_num INTEGER NOT NULL REFERENCES comics(num) ON DELETE CASCADE,
    UNIQUE (word, comics_num)
);

CREATE INDEX idx_words_word ON words(word);
CREATE INDEX idx_words_num ON words(comics_num);
