package users

import (
	"avitotask/pkg/common/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type ReserveBalanceAndRevenueRecognitionRequestBody struct {
	UserId    int `json:"user_id"`
	ServiceId int `json:"service_id"`
	OrderId   int `json:"order_id"`
	Cost      int `json:"cost"`
}

func (h handler) ReserveBalanceAndRevenueRecognition(c *gin.Context) {

	body := ReserveBalanceAndRevenueRecognitionRequestBody{}

	// получаем тело запроса
	if err := c.BindJSON(&body); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	var user models.User
	var transaction models.Transaction
	var reservation models.Reservation
	var revenue models.Revenue
	if result := h.DB.First(&user, body.UserId); result.Error != nil {
		c.AbortWithError(http.StatusBadRequest, result.Error)
		c.JSON(http.StatusBadRequest, "У заказчика с таким ID отсутствует баланс")
		return
	}
	if body.Cost <= 0 || body.ServiceId <= 0 || body.OrderId <= 0 {
		c.JSON(http.StatusBadRequest, "Сумма покупки, ID услуги, ID заказа не могут быть меньше либо равны нулю")
		return
	}
	if result := h.DB.First(&reservation, body.OrderId); result.Error != nil {

		reservation.UserId = body.UserId
		reservation.ServiceId = body.ServiceId
		reservation.OrderId = body.OrderId
		reservation.Cost = body.Cost
		reservation.Status = "Заказ не подтвержден"
		h.DB.Save(&reservation)
		c.JSON(http.StatusOK, &user)
		return
	}

	if reservation.UserId != body.UserId {
		c.JSON(http.StatusBadRequest, "Данный заказ принадлежит другому пользователю")
		return
	}
	if reservation.Status == "Заказ подтвержден" {
		c.JSON(http.StatusBadRequest, "Невозможно подтвердить уже оплаченный заказ")
		return
	}
	if reservation.ServiceId != body.ServiceId {
		c.JSON(http.StatusBadRequest, "Усулга резервации не совпадает с услугой подтверждения")
		return
	}

	if reservation.Cost != body.Cost {
		c.JSON(http.StatusBadRequest, "Сумма заказа не совпадает с суммой резервации")
		return
	}
	if user.Balance < reservation.Cost {
		c.JSON(http.StatusBadRequest, "Недостаточно средств на балнсе для подтверждения заказа")
		return
	}
	user.Balance -= reservation.Cost
	h.DB.Save(&user)
	reservation.Status = "Заказ подтвержден"
	h.DB.Save(&reservation)
	transaction.UserId = reservation.UserId
	transaction.Description = "Подтверждение заказа: " + strconv.Itoa(reservation.OrderId) + " на сумму " + strconv.Itoa(reservation.Cost) + " копеек"
	transaction.Amount = reservation.Cost
	h.DB.Save(&transaction)
	revenue.UserId = reservation.UserId
	revenue.ServiceId = reservation.ServiceId
	revenue.Amount = reservation.Cost
	revenue.OrderId = reservation.OrderId
	h.DB.Save(&revenue)
	c.JSON(http.StatusOK, &user)

}
