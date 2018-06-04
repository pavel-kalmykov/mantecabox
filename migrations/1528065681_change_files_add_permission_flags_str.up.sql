ALTER TABLE files
  ADD permissions_str char(9) DEFAULT 'rw-r--r--' NOT NULL;
