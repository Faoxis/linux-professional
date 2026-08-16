# Многоступенчатая сборка: maven нужен только на этапе компиляции,
# в финальный образ попадает jar и JRE, без тулчейна и кеша зависимостей.
FROM maven:3.9-eclipse-temurin-17 AS build
WORKDIR /src
# Сначала только pom: слой с зависимостями переиспользуется,
# пока pom не меняется, и пересборка кода их не перекачивает
COPY pom.xml ./
RUN mvn -q -B dependency:go-offline
COPY src ./src
RUN mvn -q -B package -DskipTests

FROM eclipse-temurin:17-jre-jammy
RUN useradd -u 10001 -m app
COPY --from=build /src/target/web-1.0.0.jar /app/app.jar
USER app
EXPOSE 8080
ENTRYPOINT ["java", "-jar", "/app/app.jar"]
