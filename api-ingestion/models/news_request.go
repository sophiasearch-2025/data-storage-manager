package models

type NewsRequest struct {
	URL            string   `json:"url" binding:"required"`
	Titulo         string   `json:"titulo" binding:"required"`
	Fecha          string   `json:"fecha" binding:"required"`
	Tags           []string `json:"tags"`
	Autor          string   `json:"autor"`
	DescAutor      string   `json:"desc_autor"`
	Abstract       string   `json:"abstract"`
	Cuerpo         string   `json:"cuerpo" binding:"required"`
	Multimedia     []string `json:"multimedia"`
	TipoMultimedia string   `json:"tipo_multimedia"`
}

type NewsResponse struct {
	JobID   string `json:"job_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}
