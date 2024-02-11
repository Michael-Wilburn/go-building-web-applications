CREATE TABLE `pages` (
 `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
 `page_guid` varchar(256) NOT NULL DEFAULT '',
 `page_title` varchar(256) DEFAULT NULL,
 `page_content` mediumtext,
 `page_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON
UPDATE CURRENT_TIMESTAMP,
 PRIMARY KEY (`id`),
 UNIQUE KEY `page_guid` (`page_guid`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=latin1;

INSERT INTO `pages` (`id`, `page_guid`, `page_title`,
`page_content`, `page_date`) VALUES (NULL, 'hello-world', 'Hello,
World', 'I\'m so glad you found this page! It\'s been sitting
patiently on the Internet for some time, just waiting for a
visitor.', CURRENT_TIMESTAMP);
