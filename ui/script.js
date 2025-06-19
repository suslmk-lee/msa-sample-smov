// 전역 변수
let currentUsers = [];
let currentMovies = [];
let currentBookings = [];

// ID를 이름으로 변환하는 헬퍼 함수들
function getUserNameById(userId) {
    const user = currentUsers.find(u => u.id === userId);
    return user ? user.name : userId;
}

function getMovieTitleById(movieId) {
    const movie = currentMovies.find(m => m.id === movieId);
    return movie ? movie.title : movieId;
}

// 초기 데이터 생성 및 로딩
window.onload = function() {
    initializeData();
    loadUsers();
    loadMovies();
    loadBookings();
    loadDeploymentStatus();
};

async function initializeData() {
    // 기존 데이터 확인
    try {
        const usersResponse = await fetch('/users/');
        const existingUsers = await usersResponse.json();
        
        const moviesResponse = await fetch('/movies/');
        const existingMovies = await moviesResponse.json();
        
        // 이미 데이터가 있으면 초기화 건너뛰기
        if (existingUsers.length > 0 && existingMovies.length > 0) {
            console.log('기존 데이터가 있어 초기화를 건너뜁니다.');
            return;
        }
    } catch (error) {
        console.log('기존 데이터 확인 중 오류:', error);
    }

    // 초기 사용자 데이터 생성
    const users = [
        { name: '김영희', email: 'kim@example.com' },
        { name: '이철수', email: 'lee@example.com' },
        { name: '박민수', email: 'park@example.com' }
    ];

    // 초기 영화 데이터 생성
    const movies = [
        { title: '어벤져스: 엔드게임', genre: '액션', year: 2019 },
        { title: '기생충', genre: '드라마', year: 2019 },
        { title: '탑건: 매버릭', genre: '액션', year: 2022 }
    ];

    // 사용자 데이터 초기화
    for (const user of users) {
        try {
            await fetch('/users/', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(user)
            });
        } catch (error) {
            console.log('사용자 데이터 초기화 중 오류:', error);
        }
    }

    // 영화 데이터 초기화
    for (const movie of movies) {
        try {
            await fetch('/movies/', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(movie)
            });
        } catch (error) {
            console.log('영화 데이터 초기화 중 오류:', error);
        }
    }
}

async function loadUsers() {
    try {
        const response = await fetch('/users/');
        const users = await response.json();
        currentUsers = users; // 전역 변수에 저장
        displayUsers(users);
    } catch (error) {
        document.getElementById('users').innerHTML = '<p style="color: red;">사용자 데이터 로딩 실패</p>';
    }
}

async function loadMovies() {
    try {
        const response = await fetch('/movies/');
        const movies = await response.json();
        currentMovies = movies; // 전역 변수에 저장
        displayMovies(movies);
    } catch (error) {
        document.getElementById('movies').innerHTML = '<p style="color: red;">영화 데이터 로딩 실패</p>';
    }
}

async function loadBookings() {
    try {
        const response = await fetch('/bookings/');
        const bookings = await response.json();
        currentBookings = bookings; // 전역 변수에 저장
        displayBookings(bookings);
    } catch (error) {
        document.getElementById('bookings').innerHTML = '<p style="color: red;">예약 데이터 로딩 실패</p>';
    }
}

function displayUsers(users) {
    const userDiv = document.getElementById('users');
    if (users && users.length > 0) {
        userDiv.innerHTML = users.map(user => 
            `<div class="data-item">
                <strong>${user.name}</strong><br>
                이메일: ${user.email}<br>
                ID: ${user.id || 'N/A'}
            </div>`
        ).join('');
    } else {
        userDiv.innerHTML = '<p>등록된 사용자가 없습니다.</p>';
    }
}

function displayMovies(movies) {
    const movieDiv = document.getElementById('movies');
    if (movies && movies.length > 0) {
        movieDiv.innerHTML = movies.map(movie => 
            `<div class="data-item" data-movie-id="${movie.id}">
                <strong>${movie.title}</strong><br>
                장르: ${movie.genre}<br>
                년도: ${movie.year}<br>
                <button onclick="bookMovie('${movie.id}', '${movie.title}')" class="book-btn">예약하기</button>
            </div>`
        ).join('');
    } else {
        movieDiv.innerHTML = '<p>등록된 영화가 없습니다.</p>';
    }
}

function displayBookings(bookings) {
    const bookingDiv = document.getElementById('bookings');
    if (bookings && bookings.length > 0) {
        bookingDiv.innerHTML = bookings.map(booking => 
            `<div class="data-item">
                예약 ID: ${booking.id || 'N/A'}<br>
                사용자: ${getUserNameById(booking.userId)}<br>
                영화: ${getMovieTitleById(booking.movieId)}<br>
                좌석: ${Array.isArray(booking.seats) ? booking.seats.join(', ') : (booking.seats || 'N/A')}<br>
                예약일: ${booking.bookingDate || new Date().toLocaleDateString('ko-KR')}
            </div>`
        ).join('');
    } else {
        bookingDiv.innerHTML = '<p>예약 내역이 없습니다.</p>';
    }
}

async function bookMovie(movieId, movieTitle) {
    try {
        // 간단한 예약 데이터 생성
        const numSeats = Math.floor(Math.random() * 4) + 1; // 1-4석 랜덤
        const seats = [];
        for (let i = 0; i < numSeats; i++) {
            const row = String.fromCharCode(65 + Math.floor(Math.random() * 5)); // A-E
            const num = Math.floor(Math.random() * 10) + 1; // 1-10
            seats.push(`${row}${num}`);
        }
        
        // 첫 번째 사용자를 기본 사용자로 사용 (실제로는 로그인 시스템 필요)
        const defaultUserId = currentUsers.length > 0 ? currentUsers[0].id : 'user-1';
        
        const bookingData = {
            userId: defaultUserId,
            movieId: movieId,
            seats: seats
        };

        const response = await fetch('/bookings/', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(bookingData)
        });

        if (response.ok) {
            alert(`"${movieTitle}" 영화 예약이 완료되었습니다!`);
            loadBookings(); // 예약 목록 새로고침
        } else {
            alert('예약 처리 중 오류가 발생했습니다.');
        }
    } catch (error) {
        alert('예약 요청 실패: ' + error.message);
    }
}

// 멀티클라우드 배포 상태 표시
async function loadDeploymentStatus() {
    const deploymentDiv = document.getElementById('deployment-status');
    
    try {
        const response = await fetch('/deployment-status');
        const deploymentInfo = await response.json();
        
        displayDeploymentStatus(deploymentInfo);
    } catch (error) {
        console.error('배포 상태 로딩 실패:', error);
        deploymentDiv.innerHTML = '<p style="color: red;">배포 상태 데이터 로딩 실패</p>';
    }
}

function displayDeploymentStatus(deploymentInfo) {
    const deploymentDiv = document.getElementById('deployment-status');
    
    if (!deploymentInfo || deploymentInfo.length === 0) {
        deploymentDiv.innerHTML = '<p>배포 상태 정보가 없습니다.</p>';
        return;
    }

    deploymentDiv.innerHTML = `
        <div class="deployment-grid">
            ${deploymentInfo.map(info => `
                <div class="deployment-item ${info.status === '운영중' ? 'status-running' : 'status-stopped'}">
                    <div class="service-header">
                        <span class="service-icon">${info.icon}</span>
                        <h3>${info.service}</h3>
                    </div>
                    <div class="deployment-details">
                        <div class="detail-row">
                            <strong>플랫폼:</strong> ${info.platform}
                        </div>
                        <div class="detail-row">
                            <strong>환경:</strong> ${info.environment}
                        </div>
                        <div class="detail-row">
                            <strong>컨테이너 ID:</strong> ${info.containerID || 'N/A'}
                        </div>
                        <div class="detail-row">
                            <strong>포트:</strong> ${info.port}
                        </div>
                        <div class="detail-row">
                            <strong>최종 확인:</strong> ${info.lastChecked}
                        </div>
                        <div class="detail-row status-row">
                            <strong>상태:</strong> 
                            <span class="status-badge ${info.status === '운영중' ? 'badge-running' : 'badge-stopped'}">
                                ${info.status}
                            </span>
                        </div>
                    </div>
                </div>
            `).join('')}
        </div>
        <div class="deployment-summary">
            <h4>🐳 Docker Compose 배포 개요</h4>
            <div class="summary-grid">
                <div class="summary-item">
                    <strong>현재 환경</strong><br>
                    Local Development (Docker Compose)
                </div>
                <div class="summary-item">
                    <strong>실행 중인 서비스</strong><br>
                    ${deploymentInfo.filter(s => s.status === '운영중').length}개 / ${deploymentInfo.length}개
                </div>
            </div>
            <p class="deployment-note">
                <strong>🐳 컨테이너 환경:</strong> Docker Compose를 통한 로컬 개발 환경<br>
                <strong>🔗 네트워크:</strong> Docker 내부 네트워크를 통한 서비스 간 통신<br>
                <strong>⏰ 실시간 상태:</strong> Docker 컨테이너 상태 및 헬스체크 결과 반영<br>
                <strong>📝 참고:</strong> Kubernetes 배포 시 클러스터/네임스페이스 정보로 대체됩니다
            </p>
        </div>
    `;
}