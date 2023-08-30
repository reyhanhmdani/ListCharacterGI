CREATE TABLE characters
(
    id BIGINT NOT NULL AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    element ENUM('Pyro', 'Hydro', 'Geo', 'Cyro', 'Electro', 'Anemo', 'Dendro') NOT NULL,
    age INT(3),
    address ENUM('Mondstadt', 'Liyue', 'Inazuma', 'Sumeru', 'Fontaine', 'Natlan', 'Snezhnaya') NOT NULL,
    weapon_type ENUM('Claimore', 'Polearm', 'Sword', 'Catalyst', 'Bow') NOT NULL,
    star_rating ENUM('4', '5') NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
