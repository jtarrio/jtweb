CREATE TABLE `Posts` (
  `PostId` varchar(255) NOT NULL,
  `State` tinyint(1) NOT NULL,
  `StateFromWeb` tinyint(1) DEFAULT NULL,
  PRIMARY KEY (`PostId`)
) DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE `Comments` (
  `PostId` varchar(255) NOT NULL,
  `CommentId` bigint(20) NOT NULL AUTO_INCREMENT,
  `Visible` boolean NOT NULL,
  `Author` varchar(255) NOT NULL,
  `Date` datetime NOT NULL,
  `Text` text NOT NULL,
  PRIMARY KEY (`CommentId`),
  KEY `PostId` (`PostId`),
  CONSTRAINT `Comments_ibfk_1` FOREIGN KEY (`PostId`) REFERENCES `Posts` (`PostId`)
) DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
