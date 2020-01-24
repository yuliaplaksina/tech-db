--
--
SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;


CREATE EXTENSION citext;

-- forum

CREATE TABLE forum (
      id integer NOT NULL PRIMARY KEY,
      slug citext NOT NULL,
      threads integer DEFAULT 0 NOT NULL,
      posts integer DEFAULT 0 NOT NULL,
      title varchar(100) NOT NULL,
      "user" citext NOT NULL
);


ALTER TABLE forum OWNER TO postgres;

CREATE SEQUENCE forum_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE forum_id_seq OWNER TO postgres;
ALTER SEQUENCE forum_id_seq OWNED BY forum.id;


CREATE UNIQUE INDEX forum_slug_uindex ON forum USING btree (slug);


CREATE TABLE forum_user (
       forum_id integer NOT NULL,
       user_id integer NOT NULL
);


ALTER TABLE forum_user OWNER TO postgres;

-- post

CREATE TABLE post (
     id integer NOT NULL PRIMARY KEY,
     author citext NOT NULL,
     created text NOT NULL,
     forum citext NOT NULL,
     is_edited boolean DEFAULT false NOT NULL,
     message text NOT NULL,
     parent integer DEFAULT 0 NOT NULL,
     thread integer NOT NULL,
     path bigint[] DEFAULT '{0}'::bigint[] NOT NULL
);

ALTER TABLE post OWNER TO postgres;

CREATE SEQUENCE post_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER TABLE post_id_seq OWNER TO postgres;
ALTER SEQUENCE post_id_seq OWNED BY post.id;


CREATE INDEX post_author_forum_index ON post USING btree (author, forum);
CREATE INDEX post_forum_index ON post USING btree (forum);
CREATE INDEX post_parent_index ON post USING btree (parent);
CREATE INDEX post_path_index ON post USING gin (path);
CREATE INDEX post_thread_index ON post USING btree (thread);

-- thread

CREATE TABLE thread (
       id integer NOT NULL PRIMARY KEY,
       author citext NOT NULL,
       created timestamp with time zone DEFAULT now() NOT NULL,
       forum citext NOT NULL,
       message text NOT NULL,
       slug citext,
       title varchar NOT NULL,
       votes integer DEFAULT 0
);

ALTER TABLE thread OWNER TO postgres;

CREATE SEQUENCE thread_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER TABLE thread_id_seq OWNER TO postgres;
ALTER SEQUENCE thread_id_seq OWNED BY thread.id;


CREATE INDEX thread_forum_index ON thread USING btree (forum);
CREATE UNIQUE INDEX thread_id_uindex ON thread USING btree (id);
CREATE INDEX thread_slug_index ON thread USING btree (slug);

--- user

CREATE TABLE "user" (
       id integer NOT NULL PRIMARY KEY,
       nick_name citext NOT NULL,
       email citext NOT NULL,
       full_name varchar NOT NULL,
       about text
);


ALTER TABLE "user" OWNER TO postgres;

CREATE SEQUENCE user_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE user_id_seq OWNER TO postgres;
ALTER SEQUENCE user_id_seq OWNED BY "user".id;

CREATE UNIQUE INDEX user_email_uindex ON "user" USING btree (email);
CREATE UNIQUE INDEX user_nick_name_uindex ON "user" USING btree (nick_name);
CREATE INDEX user_index ON "user" USING btree (nick_name, email, full_name, about);


CREATE TABLE vote (
     user_id integer,
     voice integer NOT NULL,
     thread_id integer NOT NULL
);


ALTER TABLE vote OWNER TO postgres;

----

ALTER TABLE ONLY forum ALTER COLUMN id SET DEFAULT nextval('forum_id_seq'::regclass);

ALTER TABLE ONLY post ALTER COLUMN id SET DEFAULT nextval('post_id_seq'::regclass);

ALTER TABLE ONLY thread ALTER COLUMN id SET DEFAULT nextval('thread_id_seq'::regclass);

ALTER TABLE ONLY "user" ALTER COLUMN id SET DEFAULT nextval('user_id_seq'::regclass);

-------

SELECT pg_catalog.setval('forum_id_seq', 1, false);

SELECT pg_catalog.setval('post_id_seq', 1, false);

SELECT pg_catalog.setval('thread_id_seq', 1, false);

SELECT pg_catalog.setval('user_id_seq', 1, false);

------
