// Обработка формы для отелей
document.getElementById('hotelForm')?.addEventListener('submit', function (e) {
    e.preventDefault();
    const destination = document.getElementById('hotel-destination').value;
    const checkIn = document.getElementById('hotel-checkin').value;
    const checkOut = document.getElementById('hotel-checkout').value;
    const guests = document.getElementById('hotel-guests').value;

    alert(`Searching hotels in ${destination} from ${checkIn} to ${checkOut} for ${guests} guests.`);
});

// Обработка формы для машин
document.getElementById('carForm')?.addEventListener('submit', function (e) {
    e.preventDefault();
    const pickup = document.getElementById('car-pickup').value;
    const pickupDate = document.getElementById('car-pickup-date').value;
    const dropoffDate = document.getElementById('car-dropoff-date').value;
    const carType = document.getElementById('car-type').value;

    alert(`Searching ${carType} cars for pickup at ${pickup} from ${pickupDate} to ${dropoffDate}.`);
});

// Обработка формы "Contact Us"
document.getElementById('contactForm')?.addEventListener('submit', async function (e) {
    e.preventDefault();

    const name = document.getElementById('name').value;
    const email = document.getElementById('email').value;
    const message = document.getElementById('message').value;

    if (!name || !email || !message) {
        alert("All fields are required.");
        return;
    }

    try {
        const response = await fetch('/contact', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ message: `${name}: ${email} - ${message}` }),
        });

        const result = await response.json();
        document.getElementById('response').innerText = JSON.stringify(result, null, 2);
    } catch (error) {
        console.error("Error submitting form:", error);
        document.getElementById('response').innerText = "Error submitting data.";
    }
});

// Динамическое отображение отелей
const hotels = [
    { name: "Luxury Inn", price: 200, rating: 4.5 },
    { name: "Economy Stay", price: 50, rating: 3.5 },
    { name: "Comfort Suites", price: 100, rating: 4.0 },
];

const hotelList = document.getElementById("hotelList");
const hotelSortSelect = document.getElementById("sortHotels");

const displayHotels = (hotels) => {
    if (hotelList) {
        hotelList.innerHTML = "";
        hotels.forEach(hotel => {
            hotelList.innerHTML += `
                <div class="hotel">
                    <h3>${hotel.name}</h3>
                    <p>Price: $${hotel.price} / night</p>
                    <p>Rating: ${hotel.rating} stars</p>
                </div>
            `;
        });
    }
};

hotelSortSelect?.addEventListener("change", () => {
    const criteria = hotelSortSelect.value;
    const sortedHotels = [...hotels];
    if (criteria === "price") sortedHotels.sort((a, b) => a.price - b.price);
    else if (criteria === "rating") sortedHotels.sort((a, b) => b.rating - a.rating);
    else if (criteria === "name") sortedHotels.sort((a, b) => a.name.localeCompare(b.name));
    displayHotels(sortedHotels);
});

displayHotels(hotels);

// Динамическое отображение автомобилей
const cars = [
    { model: "Toyota Corolla", price: 50, rating: 4.5 },
    { model: "Ford Explorer", price: 80, rating: 4.0 },
    { model: "Tesla Model 3", price: 120, rating: 5.0 },
    { model: "Honda Civic", price: 40, rating: 4.2 },
];

const carList = document.getElementById("carList");
const carSortSelect = document.getElementById("sortCars");

const displayCars = (cars) => {
    if (carList) {
        carList.innerHTML = "";
        cars.forEach(car => {
            carList.innerHTML += `
                <div class="car">
                    <h3>${car.model}</h3>
                    <p>Price: $${car.price} / day</p>
                    <p>Rating: ${car.rating} stars</p>
                </div>
            `;
        });
    }
};

carSortSelect?.addEventListener("change", () => {
    const criteria = carSortSelect.value;
    const sortedCars = [...cars];
    if (criteria === "price") sortedCars.sort((a, b) => a.price - b.price);
    else if (criteria === "rating") sortedCars.sort((a, b) => b.rating - a.rating);
    else if (criteria === "model") sortedCars.sort((a, b) => a.model.localeCompare(b.model));
    displayCars(sortedCars);
});

displayCars(cars);

document.getElementById('loginForm').addEventListener('submit', function(e) {
    e.preventDefault();
    
    let email = document.getElementById('email').value;
    let password = document.getElementById('password').value;

    fetch('/login', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ email: email, password: password })
    })
    .then(response => response.json())
    .then(data => {
        if (data.status === "success") {
            window.location.href = '/profile'; // Перенаправление на страницу профиля
        } else {
            alert(data.message); // Показать сообщение об ошибке
        }
    })
    .catch(error => {
        console.error('Ошибка:', error);
    });
});

