INSERT OR IGNORE INTO entities (templateID,hp,locationID)
VALUES
(11,99999,7),
(12,99999,4),
(13,99999,3),
(14,99999,22),
(15,99999,37);

INSERT OR IGNORE INTO merchants (entityID,sellRate,buyRate)
SELECT id,1.0,1.2 FROM entities WHERE templateID=11;
INSERT OR IGNORE INTO merchants (entityID,sellRate,buyRate)
SELECT id,5.0,10.0 FROM entities WHERE templateID=12;
INSERT OR IGNORE INTO merchants (entityID,sellRate,buyRate)
SELECT id,7.0,15.0 FROM entities WHERE templateID=13;
INSERT OR IGNORE INTO merchants (entityID,sellRate,buyRate)
SELECT id,3.0,6.0 FROM entities WHERE templateID=14;
INSERT OR IGNORE INTO merchants (entityID,sellRate,buyRate)
SELECT id,1.1,1.7 FROM entities WHERE templateID=15;

INSERT OR IGNORE INTO merchant_list
SELECT id,0 FROM entities WHERE templateID=11;
INSERT OR IGNORE INTO merchant_list
SELECT id,1 FROM entities WHERE templateID=11;
INSERT OR IGNORE INTO merchant_list
SELECT id,2 FROM entities WHERE templateID=11;
INSERT OR IGNORE INTO merchant_list
SELECT id,5 FROM entities WHERE templateID=13;
INSERT OR IGNORE INTO merchant_list
SELECT id,6 FROM entities WHERE templateID=12;
INSERT OR IGNORE INTO merchant_list
SELECT id,7 FROM entities WHERE templateID=14;

INSERT OR IGNORE INTO merchant_list (merchantID,templateID) SELECT id,15 FROM entities WHERE templateID=15;
INSERT OR IGNORE INTO merchant_list (merchantID,templateID) SELECT id,16 FROM entities WHERE templateID=15;
INSERT OR IGNORE INTO merchant_list (merchantID,templateID) SELECT id,17 FROM entities WHERE templateID=15;
INSERT OR IGNORE INTO merchant_list (merchantID,templateID) SELECT id,18 FROM entities WHERE templateID=15;
