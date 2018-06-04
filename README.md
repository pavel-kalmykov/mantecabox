# Mantecabox
![if_key_1055040 1](https://user-images.githubusercontent.com/9091619/40943992-6cf2cb0c-6853-11e8-8116-c5405bc97486.png)

Servicio de almacenamiento de ficheros seguro en la nube para la asignatura de Seguridad en el Diseño de Software

Desarrollada por [Raúl Pera Pairó](github.com/rpairo) y [Pavel Razgovorov](github.com/paveltrufi)

## Resumen
En esta práctica hemos desarrollado un sistema de backup de ficheros remoto que permite el almacenamiento y recuperación de forma segura. Funciona como una arquitectura cliente-servidor convencional. A lo largo de este documento explicaremos el diseño elegido, el proceso de desarrollo seguido, el flujo de trabajo empleado y algunos defectos y mejoras que proponemos en el hipotético caso del seguimiento del desarrollo de este proyecto.

## Diseño de la arquitectura
### Base de datos
Se trata de una arquitectura en la que un cliente es capaz de listar, subir y descargar ficheros a un servidor que almacenará los mismos. Para poder realizar estas operaciones, el cliente antes debe de registrarse en el servicio, así como obtener un permiso de inicio de sesión único.

El servidor administra tanto los datos de usuario como los metadatos del fichero en una base de datos PostgreSQL a la cual se accede de forma segura utilizando un certificado SSL. Para la administración de dicha base de datos, utilizamos el contenedor oficial de [postgres](https://hub.docker.com/_/postgres/). Hemos creado un [script en bash](https://github.com/paveltrufi/mantecabox/blob/master/init-docker-postgres-db.sh) para arrancar dicho contenedor, antes parando la instancia del mismo en caso de que ya se estuviera ejecutando.

Para mantener un estado consistente de la base de datos entre los diferentes entornos de desarrollo (varios por cada desarrollador), hemos utilizado un [sistema de migraciones de bases de datos](https://github.com/golang-migrate/migrate). La herramienta se presenta tanto en formato CLI para poder crear y ejecutar las migraciones desde el terminal, así como una librería en Go la cual hemos integrado en nuestro proyecto para que ejecute todas las migraciones automáticamente antes de arrancar el servidor; de esta forma, podemos tener un control totalmente automatizado del estado del esquema de la base de datos.

Respecto al esquema utilizado, destacamos su sencillez ya que solamente contiene tres relaciones sin contar la de control de versión de migraciones. Simplemente almacenamos la información de los usuarios y contemplamos que un usuario tiene muchos ficheros y muchos intentos de login (esto último se explicará en la parte de implementación).

![db](https://user-images.githubusercontent.com/9091619/40944723-1f321d7a-6856-11e8-9fc8-5ad879860719.png)

### Autenticación de usuarios
Para la autenticación de usuarios, hemos implementado un flow basado en el actualmente usado por [Dropbox](https://blogs.dropbox.com/tech/2016/09/how-dropbox-securely-stores-your-passwords/). Con él, el servidor espera un correo y una contraseña hasheada mediante el algoritmo SHA-512 (así lo requiere el mismo) y, una vez recibidas las credenciales, éste vuelve a hashear dicha contraseña haciendo uso del algoritmo bcrypt como salt por usuario para finalmente encriptarlo con cifrado AES en modo CTR haciendo uso de una clave que lee de un fichero de configuración que usa el servidor para su funcionamiento antes de persistir el registro en la base de datos.

Para el login hemos implementado un sistema de accesos seguros mediente tokens JWT. Para poder recibir dicho token, empleamos un factor de autenticación en dos pasos haciendo que el usuario tenga que introducir una clave de seis dígitos que recibe por correo para así asegurarnos de que el acceso va a ser legítimo. Cabe destacar también que el sistema registra los intentos de login para poder establecer un bloqueo de inicio de sesión durante un tiempo tras varios intentos (parámetros configurables), además de envíar un reporte al correo del usuario en caso de detección de actividad sospechosa o nuevos dispositivos desde donde se ha intentado iniciar sesión.

Desde la parte del cliente, además, utilizamos el algoritmo [zxcvbn](https://github.com/nbutton23/zxcvbn-go), implementado por Dropbox, para determinar la fuerza de la contraseña al registrarse forzando a que esta cumpla con un mínimo de seguridad. Por la parte del login, una vez obtenido el token JWT éste lo almacena de forma segura en el keyring del sistema operativo para que el mismo cliente pueda acceder a él más adelante.

### Persistencia de ficheros
Para el almacenamiento de ficheros, hemos implementado un almacenamiento de esquema simple con cifrado AES de tipo CTR en el servidor y además un sistema de versiones. Cada vez que el servidor recibe un fichero de un usuario, guarda sus metadatos, cifra el binario recibido, persiste el binario cifrado (desde la configuración podemos elegir si es local o remoto en Google Drive) y devuelve los metadatos del cliente.

Desde dicho cliente podremos, además de subir ficheros al servicio, listarlos, descargarlos (ya sea la última versión o eligiendo una especifica) o incluso realizar una sincronización de ficheros de forma que cada nuevo fichero que se almacene en la carpeta del cliente sea subido automáticamente al servicio.
