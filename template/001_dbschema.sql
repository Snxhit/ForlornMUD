CREATE TABLE IF NOT EXISTS spawners (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    locationID INT,
    templateType TEXT,
    templateID INT,
    duration INT,
    maxSpawns INT
);

CREATE TABLE IF NOT EXISTS rooms (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT,
    description TEXT,
    n INT, s INT, w INT, e INT
);

CREATE TABLE IF NOT EXISTS players (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    hp INT,
    str INT, dex INT, agi INT, stam INT, int INT,
    exp INT,
    level INT,
    trains INT,
    maxHp INT,
    coins INT,
    locationID INT
);

CREATE TABLE IF NOT EXISTS item_templates (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    itype TEXT,
    baseDam INT,
    baseDef INT,
    baseValue INT
);

CREATE TABLE IF NOT EXISTS item_template_modifiers (
    sourceID INT,
    stat TEXT,
    value INT
);

CREATE TABLE IF NOT EXISTS item_template_effects (
    sourceID INT,
    effect TEXT,
    value INT
);

CREATE TABLE IF NOT EXISTS items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    templateID INT,
    locationType TEXT,
    locationID INT,
    equipped INT
);

CREATE TABLE IF NOT EXISTS entity_templates (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    str INT, dex INT, agi INT, stam INT, int INT,
    level INT, aggro INT,
    maxHp INT,
    baseDam INT, baseDef INT,
    baseExp INT,
    cMin INT, cMax INT
);

CREATE TABLE IF NOT EXISTS entity_template_drops (
    entityTemplateID INT,
    itemTemplateID INT,
    chance INT,
    min INT,
    max INT
);

CREATE TABLE IF NOT EXISTS entities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    templateID INT,
    hp INT,
    locationID INT
);

CREATE TABLE IF NOT EXISTS merchants (
    entityID INT PRIMARY KEY,
    sellRate REAL,
    buyRate REAL
);

CREATE TABLE IF NOT EXISTS merchant_list (
    merchantID INT,
    templateID INT
);
