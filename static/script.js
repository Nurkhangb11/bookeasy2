// Обработка формы для отелей
document.getElementById('hotelForm').addEventListener('submit', function(e) {
    e.preventDefault();
    const destination = document.getElementById('hotel-destination').value;
    const checkIn = document.getElementById('hotel-checkin').value;
    const checkOut = document.getElementById('hotel-checkout').value;
    const guests = document.getElementById('hotel-guests').value;

    alert(`Searching hotels in ${destination} from ${checkIn} to ${checkOut} for ${guests} guests.`);
});

// Обработка формы для машин
document.getElementById('carForm').addEventListener('submit', function(e) {
    e.preventDefault();
    const pickup = document.getElementById('car-pickup').value;
    const pickupDate = document.getElementById('car-pickup-date').value;
    const dropoffDate = document.getElementById('car-dropoff-date').value;
    const carType = document.getElementById('car-type').value;

    alert(`Searching ${carType} cars for pickup at ${pickup} from ${pickupDate} to ${dropoffDate}.`);
});

// Обработка формы "Contact Us" с отправкой на Go-сервер
document.getElementById('contactForm').addEventListener('submit', async function(e) {
    e.preventDefault();

    const name = document.getElementById('name').value;
    const email = document.getElementById('email').value;
    const message = document.getElementById('message').value;

    // Проверка на пустые поля
    if (!name || !email || !message) {
        alert("Все поля формы должны быть заполнены.");
        return;
    }

    try {
        const response = await fetch('/contact', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                message: `${name}: ${email} - ${message}`
            })
        });

        const result = await response.json();
        document.getElementById('response').innerText = JSON.stringify(result, null, 2);
    } catch (error) {
        console.error("Ошибка отправки данных:", error);
        document.getElementById('response').innerText = "Произошла ошибка отправки данных.";
    }
});