INSERT OR IGNORE INTO item_templates VALUES
(0,'Cookie','A classic Flavortown cookie. Crunchy, sweet, and hard to get.','consumable',0,0,1),
(1,'Chocolate Cookie','A cookie with gooey chocolate chips. Now Im hungry.','consumable',0,0,1),
(2,'Spicy Cookie','A cookie with a fiery kick.','consumable',0,0,1),
(3,'Rainbow Cookie','A cookie that shimmers with all the colors of VIBGYOR.','consumable',0,0,1),
(4,'Raspberry Pi 5','A mysterious device from another world. Its said to be able to power a whole city..','item',0,0,1),
(5,'AI Credits','A shimmering token that glows with... Artificial Intelligence?','item',0,0,1),
(6,'PCB Credits','A green chip that glows with circuitry.','item',0,0,1),
(7,'Free Domain Name','A scroll that grants you a name in the digital wonderland of the world wide web.','item',0,0,1),
(8,'Framework Laptop','A computer that flickers between realities.','mainhand',15,5,1),
(9,'Flavor Chef Hat','A tall hat that radiates confidence.','head',0,10,1),
(10,'Spatula of Destiny','A spatula said to flip the fate of any dish or project.','mainhand',20,0,1),
(11,'Slack Ring','A badge that lets you speak in mysterious channels.','ring',0,5,1),
(12,'Flavor Shield','A shield made from hardened caramel.','offhand',0,20,1),
(13,'Cookie Bag','A bag that magically produces cookies. Maybe.','consumable',0,0,1),
(14,'Energy Drink','A can that fizzes with late night coding sessions.','consumable',0,0,1);

INSERT OR IGNORE INTO item_template_modifiers VALUES
(8,'str',5),
(8,'agi',3),
(9,'int',2),
(10,'str',7),
(12,'def',10),
(11,'int',3);

INSERT OR IGNORE INTO item_template_effects VALUES
(0,'hp',10),
(1,'hp',15),
(2,'str',2),
(3,'agi',3),
(13,'hp',5),
(14,'agi',5);
