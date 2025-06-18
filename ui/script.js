document.addEventListener('DOMContentLoaded', function() {
    // DOM Elements
    const userForm = document.getElementById('user-form');
    const userInfo = document.getElementById('user-info');
    const movieList = document.getElementById('movie-list');
    const bookingModal = document.getElementById('booking-modal');
    const bookingForm = document.getElementById('booking-form');
    const bookingMovieTitle = document.getElementById('booking-movie-title');
    const bookingMovieIdInput = document.getElementById('booking-movie-id');
    const bookingResult = document.getElementById('booking-result');
    const closeModalButton = document.querySelector('.close-button');
    const viewBookingsForm = document.getElementById('view-bookings-form');
    const bookingList = document.getElementById('booking-list');

    let moviesMap = new Map();

    // Fetch initial movie list
    fetchMovies();

    // Event Listeners
    userForm.addEventListener('submit', createUser);
    bookingForm.addEventListener('submit', createBooking);
    viewBookingsForm.addEventListener('submit', viewMyBookings);
    closeModalButton.addEventListener('click', () => bookingModal.style.display = 'none');
    window.addEventListener('click', (event) => {
        if (event.target == bookingModal) {
            bookingModal.style.display = 'none';
        }
    });

    // --- Functions ---

    function fetchMovies() {
        fetch('/movies/')
            .then(response => response.json())
            .then(movies => {
                movieList.innerHTML = ''; // Clear existing list
                if (!movies) return;

                moviesMap.clear();
                movies.forEach(movie => {
                    moviesMap.set(movie.id, movie.title);

                    const li = document.createElement('li');
                    li.className = 'movie-item';
                    li.innerHTML = `
                        <div>
                            <h3>${movie.title}</h3>
                            <p>Director: ${movie.director}</p>
                        </div>
                        <button class="book-now-btn" data-movie-id="${movie.id}" data-movie-title="${movie.title}">Book Now</button>
                    `;
                    movieList.appendChild(li);
                });

                // Add event listeners to new buttons
                document.querySelectorAll('.book-now-btn').forEach(button => {
                    button.addEventListener('click', openBookingModal);
                });
            })
            .catch(error => console.error('Error fetching movies:', error));
    }

    function createUser(e) {
        e.preventDefault();
        const name = document.getElementById('user-name').value;
        const email = document.getElementById('user-email').value;

        fetch('/users/', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name, email })
        })
        .then(response => response.json())
        .then(user => {
            userInfo.innerHTML = `<p>User created! Your ID is: <strong>${user.id}</strong>. Use this ID to book movies.</p>`;
            userForm.reset();
        })
        .catch(error => {
            userInfo.innerHTML = `<p style="color: red;">Error creating user.</p>`;
            console.error('Error creating user:', error);
        });
    }

    function openBookingModal(e) {
        const movieId = e.target.getAttribute('data-movie-id');
        const movieTitle = e.target.getAttribute('data-movie-title');

        bookingMovieTitle.textContent = `Book: ${movieTitle}`;
        bookingMovieIdInput.value = movieId;
        bookingResult.innerHTML = '';
        bookingForm.reset();

        bookingModal.style.display = 'block';
    }

    function createBooking(e) {
        e.preventDefault();
        const movieId = bookingMovieIdInput.value;
        const userId = document.getElementById('booking-user-id').value;
        const seats = document.getElementById('booking-seats').value.split(',').map(s => s.trim());

        fetch('/bookings/', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ userId, movieId, seats })
        })
        .then(response => {
            if (!response.ok) {
                throw new Error('Booking failed');
            }
            return response.json();
        })
        .then(booking => {
            bookingResult.innerHTML = `<p style="color: green;">Booking successful! Booking ID: ${booking.id}</p>`;
        })
        .catch(error => {
            bookingResult.innerHTML = `<p style="color: red;">Error: ${error.message}</p>`;
            console.error('Error creating booking:', error);
        });
    }
});
