-- MySQL dump 10.13  Distrib 5.6.24, for Win64 (x86_64)
--
-- Host: 127.0.0.1    Database: vennd
-- ------------------------------------------------------
-- Server version	5.6.25

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `addressmaps`
--

DROP TABLE IF EXISTS `addressmaps`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `addressmaps` (
  `rowid` bigint(20) NOT NULL AUTO_INCREMENT,
  `counterpartyPaymentAddress` varchar(200) DEFAULT NULL,
  `nativePaymentAddress` varchar(200) DEFAULT NULL,
  `externalAddress` varchar(200) DEFAULT NULL,
  `counterpartyAddress` varchar(200) DEFAULT NULL,
  `counterpartyAssetName` varchar(200) DEFAULT NULL,
  `nativeAssetName` varchar(200) DEFAULT NULL,
  `UDF1` varchar(200) DEFAULT NULL,
  `UDF2` varchar(200) DEFAULT NULL,
  `UDF3` varchar(200) DEFAULT NULL,
  `UDF4` varchar(200) DEFAULT NULL,
  `UDF5` varchar(200) DEFAULT NULL,
  PRIMARY KEY (`rowid`),
  UNIQUE KEY `addressMaps1` (`counterpartyPaymentAddress`),
  UNIQUE KEY `addressMaps2` (`nativePaymentAddress`),
  UNIQUE KEY `addressMaps3` (`externalAddress`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `addressmaps`
--

LOCK TABLES `addressmaps` WRITE;
/*!40000 ALTER TABLE `addressmaps` DISABLE KEYS */;
/*!40000 ALTER TABLE `addressmaps` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2015-12-11  3:48:14
