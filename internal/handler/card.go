func (h *CardHandler) CreateCard(c *gin.Context) {
    var req model.CreateCardRequest
    if err := c.ShouldBind(&req); err != nil { // multipart なので ShouldBind
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "VALIDATION_ERROR"})
        return
    }

    // 画像ファイルは任意
    file, header, err := c.Request.FormFile("file")
    if err != nil && err.Error() != "http: no such file" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "ファイルの読み込みエラー", "code": "INVALID_FILE"})
        return
    }
    if file != nil {
        defer file.Close()
    }

    userID := c.GetString("user_id")
    card, err := h.svc.CreateCard(c.Request.Context(), userID, &req, file, header)
    // ... エラーハンドリング
    c.JSON(http.StatusCreated, card)
}