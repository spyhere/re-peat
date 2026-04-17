package i18n

var ruStr = Strings{
	Common: Common{
		LoadingFile: "Загрузка файла...",
	},
	Generic: Generic{
		Amount:        "Количество",
		Audio:         "Аудио",
		AudioChannels: "Аудио каналы",
		Cancel:        "Отмена",
		Editor:        "Редактор",
		Length:        "Длина",
		Markers:       "Маркера",
		Modified:      "Изменён",
		Name:          "Имя",
		Notes:         "Заметки",
		Ok:            "OK",
		Project:       "Проект",
		SampleRate:    "Частота сэмплов",
		Save:          "Сохранить",
		SaveAs:        "Сохранить как",
		Size:          "Размер",
		Tags:          "Категории",
		Time:          "Время",
		WithComments:  "С комментариями",
	},
	Markers: MarkersView{
		MCreate:            "Создать маркер",
		MDeleteALl:         "Удалить все маркера",
		MEdit:              "Редактировать маркер",
		MNamePlaceholder:   "имя маркера...",
		MNote:              "Заметки",
		SearchBPlaceholder: "поиск по имени...",
	},
	Project: ProjectView{
		MConflictLoadBody:  "Изначально эти маркера были сохранены для \"%s\", но сейчас загружен \"%s\".\nВсё еще хотите загрузить эти маркера для этого аудио файла?\n\nМаркера превышающие длину трека будут сброшены на 0 и получат категорию \"Изменён\"",
		MConflictLoadTitle: "Конфликт загрузки маркеров",
	},
}
