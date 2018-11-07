CREATE TABLE site_configuration_files (
	"id" SERIAL NOT NULL PRIMARY KEY,
    "contents" text,
    "created_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX "site_configuration_files_unique" ON site_configuration_files(id);
